package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterAllowsBurstThenBlocks(t *testing.T) {
	rl := newRateLimiter(1, 3) // 1 rps, burst of 3
	base := time.Unix(0, 0)
	rl.now = func() time.Time { return base }

	for i := 0; i < 3; i++ {
		if !rl.allow("k") {
			t.Fatalf("request %d within burst should be allowed", i+1)
		}
	}
	if rl.allow("k") {
		t.Fatal("4th request should be blocked once burst is exhausted")
	}
}

func TestRateLimiterRefills(t *testing.T) {
	rl := newRateLimiter(2, 1) // 2 tokens/sec
	now := time.Unix(0, 0)
	rl.now = func() time.Time { return now }

	if !rl.allow("k") {
		t.Fatal("first request should pass")
	}
	if rl.allow("k") {
		t.Fatal("second immediate request should be blocked")
	}
	now = now.Add(time.Second) // 2 tokens refilled, capped at burst=1
	if !rl.allow("k") {
		t.Fatal("after refill the request should pass")
	}
}

func TestRateLimiterPerKeyIsolation(t *testing.T) {
	rl := newRateLimiter(1, 1)
	rl.now = func() time.Time { return time.Unix(0, 0) }

	if !rl.allow("a") {
		t.Fatal("key a first request should pass")
	}
	if !rl.allow("b") {
		t.Fatal("key b must not be affected by key a's usage")
	}
	if rl.allow("a") {
		t.Fatal("key a should now be limited")
	}
}

func TestRateLimiterDisabled(t *testing.T) {
	rl := newRateLimiter(0, 0) // disabled
	for i := 0; i < 100; i++ {
		if !rl.allow("k") {
			t.Fatal("disabled limiter must always allow")
		}
	}
}

func TestRateLimiterMiddleware429(t *testing.T) {
	rl := newRateLimiter(1, 1)
	rl.now = func() time.Time { return time.Unix(0, 0) }
	h := rl.middleware(func(r *http.Request) string { return "same" },
		func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	rec1 := httptest.NewRecorder()
	h(rec1, httptest.NewRequest(http.MethodPost, "/search", nil))
	if rec1.Code != http.StatusOK {
		t.Fatalf("first request want 200, got %d", rec1.Code)
	}
	rec2 := httptest.NewRecorder()
	h(rec2, httptest.NewRequest(http.MethodPost, "/search", nil))
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request want 429, got %d", rec2.Code)
	}
}

func TestClientKeyPrefersAPIKey(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/search", nil)
	r.Header.Set("X-API-Key", "secret")
	r.RemoteAddr = "10.0.0.1:1234"
	if got := clientKey(r); got != "key:secret" {
		t.Fatalf("want key:secret, got %s", got)
	}
}

func TestClientKeyFallsBackToIP(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/search", nil)
	r.RemoteAddr = "10.0.0.1:1234"
	if got := clientKey(r); got != "ip:10.0.0.1" {
		t.Fatalf("want ip:10.0.0.1, got %s", got)
	}
	r.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")
	if got := clientKey(r); got != "ip:203.0.113.5" {
		t.Fatalf("want ip:203.0.113.5, got %s", got)
	}
}
