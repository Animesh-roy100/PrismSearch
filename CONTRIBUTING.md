# Contributing

Thanks for your interest in improving this project. The stack builds and tests entirely
inside containers, so the only host requirement is **Docker + Docker Compose**. A local
Go 1.26 / Python 3.12 toolchain is optional and only needed for the fast inner loop.

## Development loop

```bash
make up      # build + start the full stack, register CDC, seed data
make test    # all unit tests (Go + Python) in containers
make e2e     # end-to-end freshness + hybrid retrieval + RAG assertions
make eval    # retrieval ablation table + faithfulness
make down    # tear everything down
```

## Before opening a pull request

Run the same checks CI runs. With a local Go toolchain:

```bash
gofmt -l .              # must print nothing
go vet ./...
go build ./...
go test -race ./...
golangci-lint run ./... # https://golangci-lint.run/usage/install
govulncheck ./...       # go install golang.org/x/vuln/cmd/govulncheck@latest
```

Python:

```bash
ruff check mlservice eval
python -m pytest mlservice/test_ml.py eval/test_eval.py -q
```

All of the above run automatically in GitHub Actions on every push and pull request.

## Guidelines

- Keep the core **source-agnostic**: new sources normalize to the canonical doc shape
  in `internal/contract`; the indexer, stores, retrieval, and RAG layers stay generic.
- Add or update tests for any behavior change. Prefer deterministic tests (the `hash`
  embedding / `lexical` rerank backends need no model downloads).
- Match the existing style: small typed `config` helpers for env vars, graceful
  degradation when a dependency is down, and spans/metrics for new stages.
- Keep commits focused and write a descriptive message. Conventional-commit prefixes
  (`feat:`, `fix:`, `docs:`, `chore:`) are appreciated.

## Reporting issues

Use GitHub Issues for bugs and feature requests. For security reports, see
[SECURITY.md](SECURITY.md).
