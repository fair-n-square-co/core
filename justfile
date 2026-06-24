build:
    @echo "Building..."
    go build -o bin/core ./cmd/core
    @echo "Done."

run: build
    @echo "Running..."
    ./bin/core
    @echo "Done."

test:
    @echo "Running tests..."
    go test -race -covermode=atomic -coverprofile=cover.out -v $(go list ./... | grep -v '/mocks$')
    @echo "Done."

test-coverage: test
    @echo "Generating coverage report..."
    go tool cover -html=cover.out
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
    goose -dir db/core/migrations postgres "$DATABASE_URL" up
    @echo "Done."

generate:
    @echo "Generating sqlc..."
    sqlc generate
    @echo "Generating mocks..."
    go generate ./...
    @echo "Done."
