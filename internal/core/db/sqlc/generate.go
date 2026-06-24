package sqlc

// Mock generation for the database interfaces exercised by the generated
// queries. These directives live here rather than in db.go because that file
// is sqlc-generated (DO NOT EDIT) and would be overwritten by `sqlc generate`.
//
// Regenerate with `go generate ./...`.

// DBTX is defined in this package (db.go); generate its mock from source.
//go:generate go tool mockgen -source=db.go -destination=mocks/dbtx.go -package=mocks

// pgx.Row and pgx.Rows are third-party interfaces returned by DBTX; generate
// their mocks in reflect mode.
//go:generate go tool mockgen -destination=mocks/pgx.go -package=mocks github.com/jackc/pgx/v5 Row,Rows
