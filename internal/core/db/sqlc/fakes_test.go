package sqlc

import (
	"context"
	"reflect"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// fakeDBTX is a hand-rolled stub of the DBTX interface. Only the method a given
// query uses needs to be populated for a test.
type fakeDBTX struct {
	execFn     func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	queryFn    func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	queryRowFn func(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

func (f *fakeDBTX) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return f.execFn(ctx, sql, args...)
}

func (f *fakeDBTX) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return f.queryFn(ctx, sql, args...)
}

func (f *fakeDBTX) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return f.queryRowFn(ctx, sql, args...)
}

// fakeRow implements pgx.Row. The scan func receives the Scan destinations.
type fakeRow struct {
	scan func(dest ...any) error
}

func (r fakeRow) Scan(dest ...any) error { return r.scan(dest...) }

// fakeRows implements pgx.Rows over a fixed sequence of per-row scan funcs.
type fakeRows struct {
	scans []func(dest ...any) error
	idx   int
	err   error
}

func (r *fakeRows) Next() bool {
	if r.idx >= len(r.scans) {
		return false
	}
	r.idx++
	return true
}

func (r *fakeRows) Scan(dest ...any) error { return r.scans[r.idx-1](dest...) }
func (r *fakeRows) Close()                 {}
func (r *fakeRows) Err() error             { return r.err }

// Unused methods required to satisfy the pgx.Rows interface.
func (r *fakeRows) CommandTag() pgconn.CommandTag               { return pgconn.CommandTag{} }
func (r *fakeRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (r *fakeRows) Values() ([]any, error)                      { return nil, nil }
func (r *fakeRows) RawValues() [][]byte                         { return nil }
func (r *fakeRows) Conn() *pgx.Conn                             { return nil }

// scanValues returns a scan func that assigns each provided value into the
// matching Scan destination positionally, so tests can drive the generated
// row.Scan(&i.Field, ...) calls.
func scanValues(values ...any) func(dest ...any) error {
	return func(dest ...any) error {
		for i := range dest {
			if i >= len(values) {
				break
			}
			reflect.ValueOf(dest[i]).Elem().Set(reflect.ValueOf(values[i]))
		}
		return nil
	}
}

// scanErr returns a scan func that always fails with err.
func scanErr(err error) func(dest ...any) error {
	return func(dest ...any) error { return err }
}
