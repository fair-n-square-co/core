package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"

	"github.com/fair-n-square-co/apis/gen/pkg/fairnsquare/service/ledger/v1alpha1/ledgerpbconnect"
	"github.com/fair-n-square-co/core/cmd/core/config"
	"github.com/fair-n-square-co/core/internal/ledger/api"
	"github.com/fair-n-square-co/core/internal/ledger/repository"
	"github.com/fair-n-square-co/core/internal/ledger/service"
	"github.com/fair-n-square-co/core/pkg/middleware"
)

// shutdownTimeout bounds how long graceful shutdown waits for in-flight RPCs.
const shutdownTimeout = 10 * time.Second

// server serves the connect/gRPC API on the configured port using the given
// connection pool until ctx is cancelled. The pool is owned by the caller.
func server(ctx context.Context, cfg *config.Config, pool *pgxpool.Pool) error {
	logger := slog.Default()

	addr := fmt.Sprintf(":%d", cfg.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	srv := newHTTPServer(newMux(pool, logger))

	logger.Info("listening", "addr", addr)
	return serve(ctx, srv, lis)
}

// newMux builds the HTTP mux exposing the friend service, gRPC health, and gRPC
// reflection, all wrapped with the shared logging/recovery interceptors.
func newMux(pool *pgxpool.Pool, logger *slog.Logger) *http.ServeMux {
	interceptors := connect.WithInterceptors(
		middleware.NewRecoveryInterceptor(logger),
		middleware.NewLoggingInterceptor(logger),
	)

	friendSrv := api.NewFriendServer(service.NewFriendService(repository.New(pool)))

	mux := http.NewServeMux()
	mux.Handle(ledgerpbconnect.NewFriendServiceHandler(friendSrv, interceptors))

	checker := grpchealth.NewStaticChecker(ledgerpbconnect.FriendServiceName)
	mux.Handle(grpchealth.NewHandler(checker))

	reflector := grpcreflect.NewStaticReflector(ledgerpbconnect.FriendServiceName)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	return mux
}

// newHTTPServer builds an HTTP server that also speaks unencrypted (cleartext)
// HTTP/2, so gRPC clients can connect without TLS (local development only).
func newHTTPServer(handler http.Handler) *http.Server {
	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)

	return &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		Protocols:         protocols,
	}
}

// serve runs srv on lis and shuts it down gracefully when ctx is cancelled.
func serve(ctx context.Context, srv *http.Server, lis net.Listener) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := srv.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	})

	return g.Wait()
}
