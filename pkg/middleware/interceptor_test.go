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
