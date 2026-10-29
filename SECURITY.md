# Security Policy

## Reporting a vulnerability

Please report security issues privately rather than opening a public issue.
Open a [GitHub Security Advisory](https://github.com/Animesh-roy100/PrismSearch/security/advisories/new)
or email the maintainer. You will receive an acknowledgement within a few days.

Please include:
- A description of the issue and its impact.
- Steps to reproduce (a minimal proof of concept if possible).
- Affected version / commit.

## Supported versions

This is a demonstration / portfolio project. Security fixes are applied to `main`.

## Security model & hardening

The query API (`cmd/query`) is the only externally exposed surface. Relevant controls:

- **Authentication** — set `QUERY_API_KEY` to require an `X-API-Key` (or
  `Authorization: Bearer`) header on `/search` and `/ask`. When unset, the service runs
  in open *dev mode* and logs a warning.
- **Tenant isolation** — when an API key is configured, an `X-Tenant-ID` header is
  mandatory and the server force-injects it into every retrieval filter, so a client
  cannot read across tenants by omitting or spoofing a filter.
- **Rate limiting** — a per-client token-bucket limiter (keyed by API key, else source
  IP) guards against abuse. Tunable via `RATE_RPS` / `RATE_BURST`; set `RATE_RPS<=0` to
  disable.
- **Request size limits** — request bodies are capped at 64 KiB to prevent unbounded
  buffering.
- **Timeouts** — every request is bounded by `REQUEST_TIMEOUT`; the HTTP server sets a
  `ReadHeaderTimeout` to mitigate slowloris.
- **Dependency hygiene** — CI runs `govulncheck` and Dependabot opens weekly updates
  for Go modules, pip requirements, GitHub Actions, and Docker base images.

## Secrets

No credentials are committed to this repository. Provider API keys (e.g.
`GEMINI_API_KEY`) are read from the local `.env` (which is git-ignored) or the
environment, and consumed by the LiteLLM container — never baked into images or source.
If you fork this project, use your own keys and never commit them.
