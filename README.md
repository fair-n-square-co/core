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
│       └── config/     # app config loader + embedded config.yml / config.prod.yml
├── db/
│   └── core/
│       ├── migrations/ # goose migrations (service-level)
│       └── queries/    # sqlc SQL files
├── pkg/
│   └── middleware/     # shared middleware (deferred to FNS-87)
├── internal/
│   ├── core/
│   │   ├── db/         # pgx connection pool (NewPool) + DBConfig
│   │   │   └── sqlc/   # generated — do not edit
│   │   └── logger/     # slog setup (InitLogger)
│   └── ledger/
│       ├── api/        # gRPC handlers
│       ├── service/    # domain logic
│       └── repository/ # DB queries
├── Dockerfile
├── docker-compose.yml
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
just docker-build   # build the core image
just docker-up      # docker compose up --build (app + postgres)
just docker-down    # docker compose down
```

## Configuration

Config is resolved in `cmd/core/config` in two passes into a single struct: the
**embedded YAML** (`config.yml`, or `config.prod.yml` when `CORE_ENV=production`) is loaded
first, then **environment variables** override it. Env vars use the `CORE_` prefix with
`_` as the nested-key delimiter (`CORE_ENV` selects which YAML file loads):

| Env var | Maps to | Default |
|---------|---------|---------|
| `CORE_PORT` | `Port` | `8080` |
| `CORE_DB_CONNSTRING` | `Db.ConnString` | _(empty)_ |
| `CORE_DB_MAXCONNS` | `Db.MaxConns` | `10` (`20` in prod) |
| `CORE_DB_MINCONNS` | `Db.MinConns` | `2` (`5` in prod) |
| `CORE_DB_MAXCONNLIFETIME` | `Db.MaxConnLifetime` | `1h` |
| `CORE_DB_MAXCONNIDLETIME` | `Db.MaxConnIdleTime` | `30m` |
| `CORE_DB_HEALTHCHECKPERIOD` | `Db.HealthCheckPeriod` | `1m` |
| `CORE_LOGGER_FORMAT` | `Logger.Format` | `json` (or `text`) |
| `CORE_LOGGER_LEVEL` | `Logger.Level` | `0` (slog Info) |
| `CORE_ENV` | selects YAML file | `config.yml` |

> The app reads `CORE_DB_CONNSTRING`. `just migrate-db` is a separate goose CLI and
> still uses `DATABASE_URL`.

## Database

Core owns a **single Postgres database** (`fairnsquare_core`) and never joins across service
boundaries: it talks only to its own DB and reaches the Auth service over RPC, honoring the
ADR-2 boundary. The schema stays lean — our own ids plus external references and business data
only, no auth or profile data (see the docs repo's ADRs).

### Schema & query workflow

- **Migrations** are goose SQL files in `db/core/migrations/`, each with a `+goose Up` and a
  `+goose Down`. Generate a new one with `just migrate-gen` (it prints the `goose … create`
  command); never hand-rename existing files. Apply them with `just migrate-db` (reads
  `DATABASE_URL`).
- **Queries** are hand-written sqlc SQL in `db/core/queries/`. After editing a query or
  migration, run `just generate` to regenerate the typed Go in `internal/core/db/sqlc/` — those
  files are generated, never hand-edited.

### Connection pooling

`internal/core/db.NewPool` opens a tuned `pgxpool` and pings it so a bad DSN fails fast at
startup. The pool is wired through `config` (embedded YAML + `CORE_DB_*` env overrides) and
injected into each module's `repository/`. Tunables — `MaxConns`, `MinConns`, `MaxConnLifetime`,
`MaxConnIdleTime`, `HealthCheckPeriod` — default from the embedded config (see the table above);
any field left unset falls back to the pgx default.

## Docker

`docker compose up --build` starts Postgres and the service. The app connects via the
`CORE_DB_CONNSTRING` set in `docker-compose.yml`. The YAML config is embedded in the
binary, so nothing extra is copied into the image.
