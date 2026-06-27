# Load .env (POSTGRES_*, DATABASE_URL, ...) if present.
set dotenv-load := true

# Postgres credentials for local dev. Defaults let `just run` / `just docker-up`
# work with zero config; override any of them via .env or the environment.
# These are `export`ed so `docker compose` interpolation picks them up.
export POSTGRES_USER := env_var_or_default("POSTGRES_USER", "postgres")
export POSTGRES_PASSWORD := env_var_or_default("POSTGRES_PASSWORD", "postgres")
export POSTGRES_DB := env_var_or_default("POSTGRES_DB", "core")

# Connection string the app uses inside the compose network (host `db`).
export CORE_DB_CONNSTRING := env_var_or_default("CORE_DB_CONNSTRING", "postgres://" + POSTGRES_USER + ":" + POSTGRES_PASSWORD + "@db:5432/" + POSTGRES_DB + "?sslmode=disable")

# Connection string for the host machine talking to the compose Postgres
# (published on 127.0.0.1:5432). Overridable via DATABASE_URL.
local_db_url := env_var_or_default("DATABASE_URL", "postgres://" + POSTGRES_USER + ":" + POSTGRES_PASSWORD + "@localhost:5432/" + POSTGRES_DB + "?sslmode=disable")

build:
    @echo "Building..."
    go build -o bin/core ./cmd/core
    @echo "Done."

# Start only the Postgres service (detached) and wait for it to be healthy.
db:
    @echo "Starting Postgres..."
    docker compose up -d --wait db

# Local dev: ensure the DB is up, run migrations, then run the app with
# hot reload (air). Edits under ./cmd and ./internal rebuild automatically.
run: db
    @echo "Migrating database..."
    goose -dir db/core/migrations postgres "{{local_db_url}}" up
    @echo "Starting app with hot reload..."
    CORE_DB_CONNSTRING="{{local_db_url}}" go tool air

test:
    @echo "Running tests..."
    go test -race -covermode=atomic -coverprofile=cover.out -v $(go list ./... | grep -v '/mocks$')
    @echo "Done."

test-coverage: test
    @echo "Generating coverage report..."
    go tool cover -html=cover.out
    @echo "Done."

# Integration tests spin a real Postgres via testcontainers (requires Docker).
test-integration:
    @echo "Running integration tests..."
    go test -tags integration -race -count=1 ./cmd/core/...
    @echo "Done."

lint:
    @echo "Linting..."
    golangci-lint run ./...
    @echo "Done."

migrate-gen:
    @echo ""
    @echo "Replace <migration_name> in the command below"
    @echo "with an appropriate migration name and generate migration files"
    @echo "=============================================================="
    @echo "goose -dir db/core/migrations create <migration_name> sql"
    @echo "=============================================================="

migrate-db:
    @echo "Migrating database..."
    goose -dir db/core/migrations postgres "{{local_db_url}}" up
    @echo "Done."

docker-build: docker-build-core docker-build-migrate

docker-build-core:
    @echo "Building core image..."
    docker build -f Dockerfile.core -t core .
    @echo "Done."

docker-build-migrate:
    @echo "Building migrate image..."
    docker build -f Dockerfile.migrate -t core-migrate .
    @echo "Done."

# Verify a published image's cosign signature was produced by this repo's release workflow.
# Usage: just cosign-verify ghcr.io/fair-n-square-co/core:1.2.3
cosign-verify image:
    cosign verify {{image}} \
      --certificate-identity-regexp '^https://github.com/fair-n-square-co/core/' \
      --certificate-oidc-issuer https://token.actions.githubusercontent.com

docker-up:
    @echo "Starting services..."
    docker compose up --build

docker-down:
    @echo "Stopping services..."
    docker compose down

generate:
    @echo "Generating sqlc..."
    sqlc generate
    @echo "Generating mocks..."
    go generate ./...
    @echo "Done."
