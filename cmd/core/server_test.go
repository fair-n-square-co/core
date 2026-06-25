package main

import (
	"context"
	"crypto/tls"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"
)

// TestServe_GracefulShutdown starts serve on an ephemeral port and confirms it
// returns nil once the context is cancelled.
func TestServe_GracefulShutdown(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := newHTTPServer(http.NewServeMux())

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- serve(ctx, srv, lis) }()

	// Give the server a moment to start serving before cancelling.
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("server did not shut down within timeout")
	}
}

// TestNewHTTPServer_ServesCleartextHTTP2 exercises the production transport:
// it serves newHTTPServer over a real listener and confirms a request
// negotiates unencrypted HTTP/2 (h2c) and is routed by the connect handler.
// A nil pool is safe because a GET to a unary RPC is rejected before the
// handler runs (so the DB is never touched).
func TestNewHTTPServer_ServesCleartextHTTP2(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := newHTTPServer(newMux(nil, logger))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() { errCh <- serve(ctx, srv, lis) }()

	// An h2c client: HTTP/2 spoken over a plaintext TCP connection (no TLS).
	client := &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			},
		},
	}

	url := "http://" + lis.Addr().String() + "/fairnsquare.service.ledger.v1alpha1.FriendService/ListFriends"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, 2, resp.ProtoMajor, "expected the server to speak cleartext HTTP/2")
	assert.NotEqual(t, http.StatusNotFound, resp.StatusCode, "connect handler should route the service path")

	cancel()
	require.NoError(t, <-errCh)
}

// TestServe_PropagatesServeError confirms serve returns a non-ErrServerClosed
// error from Serve (here, a listener closed out from under it) instead of
// swallowing it.
func TestServe_PropagatesServeError(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	require.NoError(t, lis.Close()) // Serve will fail immediately on a closed listener.

	srv := newHTTPServer(http.NewServeMux())

	err = serve(context.Background(), srv, lis)
	require.Error(t, err)
}

// TestNewMux_RegistersServices verifies the mux is built and routes the friend
// service path (a nil pool is safe: the DB is only touched when an RPC runs).
func TestNewMux_RegistersServices(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mux := newMux(nil, logger)
	require.NotNil(t, mux)

	// A bare GET to the service prefix should be routed (not 404) by the
	// registered connect handler.
	rec := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/fairnsquare.service.ledger.v1alpha1.FriendService/ListFriends", nil)
	require.NoError(t, err)
	mux.ServeHTTP(rec, req)

	assert.NotEqual(t, http.StatusNotFound, rec.Code)
}
