# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- GitHub Actions CI: Go build/vet/`gofmt`/race-tests + `golangci-lint`, `govulncheck`
  vulnerability scan, and Python (`ruff` + `pytest`) jobs.
- `.golangci.yml` lint configuration.
- Per-client token-bucket rate limiting and 64 KiB request body size limits on the
  query API (`RATE_RPS`, `RATE_BURST`), with unit tests.
- `SECURITY.md`, `CONTRIBUTING.md`, and this `CHANGELOG.md`.
- Dependabot configuration for Go modules, pip, GitHub Actions, and Docker images.

### Changed
- Tightened `.gitignore` for editor/OS/tooling artifacts.

### Fixed
- Lint findings: unused struct field in the indexer, unclosed response body in the
  query healthcheck, and `nil` contexts in cache tests.

## [1.0.0] - 2026-09-27

### Added
- v1 vertical slice (Phases 0–5.5): PostgreSQL CDC via Debezium → Redpanda →
  normalizer → Go indexer → OpenSearch (BM25) + Qdrant (vectors).
- Hybrid retrieval: BM25 + vector + RRF + cross-encoder rerank + MMR.
- Grounded RAG with citations, a faithfulness guardrail, and a semantic answer cache.
- Graceful degradation across every dependency, per-source idempotency, and a DLQ.
- OpenTelemetry tracing, Prometheus metrics, and Grafana dashboards.
- Self-contained Helm chart for K3s/Kubernetes and a LiteLLM provider abstraction.
- Offline evaluation harness (nDCG@10 / MRR / Recall) and end-to-end assertions.
