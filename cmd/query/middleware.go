package main

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// maxBodyBytes caps the size of an inbound request body. Query/ask payloads are small
// JSON objects; anything larger is almost certainly abuse or a bug. http.MaxBytesReader
// makes the handler's json.Decode fail cleanly instead of buffering unbounded input.
const maxBodyBytes = 64 * 1024 // 64 KiB

// limitBody wraps a handler so each request body is bounded by maxBodyBytes.
func limitBody(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		next(w, r)
	}
}

// rateLimiter is a per-client token-bucket limiter. Clients are keyed by API key when
// present, otherwise by source IP, so a single tenant/caller cannot exhaust the service
// for everyone. Buckets are evicted after an idle window to bound memory. A refill rate
// <= 0 disables limiting entirely (open/dev mode).
type rateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rps      float64       // tokens added per second
	burst    float64       // maximum tokens (bucket capacity)
	idleEvic time.Duration // evict buckets unused for this long
	now      func() time.Time
}

type bucket struct {
	tokens float64
	last   time.Time
	seen   time.Time
}

func newRateLimiter(rps, burst float64) *rateLimiter {
	return &rateLimiter{
		buckets:  make(map[string]*bucket),
		rps:      rps,
		burst:    burst,
		idleEvic: 10 * time.Minute,
		now:      time.Now,
	}
}

// allow reports whether a request from key may proceed, consuming one token if so.
func (rl *rateLimiter) allow(key string) bool {
	if rl.rps <= 0 {
		return true
	}
	now := rl.now()
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[key]
	if !ok {
		b = &bucket{tokens: rl.burst, last: now}
		rl.buckets[key] = b
	}
	// Refill proportional to elapsed time, capped at burst.
	elapsed := now.Sub(b.last).Seconds()
	b.tokens += elapsed * rl.rps
	if b.tokens > rl.burst {
		b.tokens = rl.burst
	}
	b.last = now
	b.seen = now
	rl.evictLocked(now)

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// evictLocked drops idle buckets. Caller must hold rl.mu.
func (rl *rateLimiter) evictLocked(now time.Time) {
	for k, b := range rl.buckets {
		if now.Sub(b.seen) > rl.idleEvic {
			delete(rl.buckets, k)
		}
	}
}

// middleware applies rate limiting keyed by the provided function.
func (rl *rateLimiter) middleware(keyFn func(*http.Request) string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !rl.allow(keyFn(r)) {
			w.Header().Set("Retry-After", "1")
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate limited"})
			return
		}
		next(w, r)
	}
}

// clientKey derives a stable rate-limit key: the API key if presented, otherwise the
// request's source IP (honoring X-Forwarded-For's first hop behind a trusted proxy).
func clientKey(r *http.Request) string {
	if k := r.Header.Get("X-API-Key"); k != "" {
		return "key:" + k
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		first, _, _ := strings.Cut(xff, ",")
		return "ip:" + strings.TrimSpace(first)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	return "ip:" + host
}
