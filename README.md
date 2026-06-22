# core

Fair N Square core service — modular monolith for groups, friends, expenses, ledger, settlement, and todos.

## Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.26 | `mise install go 1.26` |
| just | latest | `brew install just` |
| sqlc | latest | `brew install sqlc` |
| goose | latest | `go install github.com/pressly/goose/v3/cmd/goose@latest` |
| golangci-lint | latest | `brew install golangci-lint` |

## Module layout

```
core/
├── cmd/
│   └── core/           # single entrypoint
├── db/
│   └── core/
│       ├── migrations/ # goose migrations (service-level)
│       └── queries/    # sqlc SQL files
├── pkg/
│   └── middleware/     # shared middleware (deferred to FNS-87)
├── internal/
│   ├── core/
│   │   ├── db/         # connection pool (deferred to FNS-87)
│   │   │   └── sqlc/   # generated — do not edit
│   │   └── config/     # app config (deferred to FNS-87)
│   └── ledger/
│       ├── api/        # gRPC handlers
│       ├── service/    # domain logic
│       └── repository/ # DB queries
├── sqlc.yml
├── justfile
└── go.mod
```

**Module layering:** `api/` → `service/` → `repository/`. No cross-module package imports; modules communicate through service interfaces. The DB connection pool is a service-level concern, injected into each `repository/`.

## Targets

```
just build          # compile to bin/core
just run            # build and run
just test           # run tests with race detector and coverage
just test-coverage  # open HTML coverage report
just lint           # run golangci-lint
just migrate-gen    # print goose migration generation command
just migrate-db     # apply goose migrations (requires $DATABASE_URL)
just generate       # run sqlc generate
```
