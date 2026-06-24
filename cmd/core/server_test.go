package main

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/fair-n-square-co/core/cmd/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServer_GracefulShutdown starts the server on an ephemeral port and
// confirms it shuts down cleanly when the context is cancelled, returning a nil
// error.
func TestServer_GracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- server(ctx, &config.Config{Port: 0})
	}()

	// Give the server a moment to bind and start serving before cancelling.
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("server did not shut down within timeout")
	}
}

// TestServer_ListenError verifies that server returns an error when the port is
// already in use.
func TestServer_ListenError(t *testing.T) {
	// Occupy a port (on all interfaces, matching how server binds) so the
	// server's Listen fails.
	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	defer func() { _ = lis.Close() }()

	_, portStr, err := net.SplitHostPort(lis.Addr().String())
	require.NoError(t, err)
	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)

	err = server(context.Background(), &config.Config{Port: port})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "listen")
}
