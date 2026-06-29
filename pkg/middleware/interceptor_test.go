package middleware

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// payload is a minimal message type; the interceptors never encode it, they
// only thread it through, so it needs no proto machinery.
type payload struct{}

func newReq() connect.AnyRequest {
	return connect.NewRequest(&payload{})
}

func TestLoggingInterceptor_LogsSuccess(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	next := connect.UnaryFunc(func(_ context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		return connect.NewResponse(&payload{}), nil
	})

	wrapped := NewLoggingInterceptor(logger).WrapUnary(next)
	resp, err := wrapped(context.Background(), newReq())

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, buf.String(), "rpc handled")
}

func TestLoggingInterceptor_LogsErrorWithCode(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	next := connect.UnaryFunc(func(_ context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("nope"))
	})

	wrapped := NewLoggingInterceptor(logger).WrapUnary(next)
	_, err := wrapped(context.Background(), newReq())

	require.Error(t, err)
	out := buf.String()
	assert.Contains(t, out, "rpc failed")
	assert.Contains(t, strings.ToLower(out), "not_found")
}

func TestRecoveryInterceptor_RecoversPanic(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	next := connect.UnaryFunc(func(_ context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		panic("boom")
	})

	wrapped := NewRecoveryInterceptor(logger).WrapUnary(next)
	resp, err := wrapped(context.Background(), newReq())

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, connect.CodeInternal, connect.CodeOf(err))
	assert.Contains(t, buf.String(), "panic recovered")
}

func TestRecoveryInterceptor_PassesThroughSuccess(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))

	next := connect.UnaryFunc(func(_ context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		return connect.NewResponse(&payload{}), nil
	})

	wrapped := NewRecoveryInterceptor(logger).WrapUnary(next)
	resp, err := wrapped(context.Background(), newReq())

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

// sanitize runs the error-sanitizer interceptor over a handler returning err and
// reports the error a client would observe. (Twin cases live in auth/pkg/middleware.)
func sanitize(t *testing.T, err error) error {
	t.Helper()
	next := connect.UnaryFunc(func(_ context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		return nil, err
	})
	wrapped := NewErrorSanitizerInterceptor().WrapUnary(next)
	_, got := wrapped(context.Background(), newReq())
	return got
}

func TestErrorSanitizer_HidesServerFaults(t *testing.T) {
	leaky := connect.NewError(connect.CodeInternal,
		errors.New(`list friendships: ERROR: relation "friendships" does not exist`))

	got := sanitize(t, leaky)

	require.Error(t, got)
	assert.Equal(t, connect.CodeInternal, connect.CodeOf(got))
	assert.Equal(t, "internal: internal error", got.Error())
	assert.NotContains(t, got.Error(), "friendships")
}

func TestErrorSanitizer_PassesClientFaultsThrough(t *testing.T) {
	for _, code := range []connect.Code{
		connect.CodeInvalidArgument,
		connect.CodeAlreadyExists,
		connect.CodeNotFound,
		connect.CodeUnauthenticated,
	} {
		orig := connect.NewError(code, errors.New("missing X-User-Id header"))
		got := sanitize(t, orig)

		require.Error(t, got)
		assert.Equal(t, code, connect.CodeOf(got))
		assert.Contains(t, got.Error(), "missing X-User-Id header")
	}
}

func TestErrorSanitizer_PassesSuccessThrough(t *testing.T) {
	assert.NoError(t, sanitize(t, nil))
}
