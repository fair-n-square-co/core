package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"

	ledgerpb "github.com/fair-n-square-co/apis/gen/pkg/fairnsquare/service/ledger/v1alpha1"
	"github.com/fair-n-square-co/core/internal/ledger/api"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func server(ctx context.Context) error {
	port := ":8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = ":" + envPort
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("listen %s: %w", port, err)
	}

	// TODO: insecure transport is for local development only. Replace with TLS
	// credentials (e.g. credentials.NewServerTLSFromFile) before production.
	grpcServer := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))
	reflection.Register(grpcServer)

	ledgerpb.RegisterFriendServiceServer(grpcServer, api.NewFriendServer())

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		slog.Info("listening", "port", port)
		return grpcServer.Serve(lis)
	})

	g.Go(func() error {
		<-ctx.Done()
		grpcServer.GracefulStop()
		return nil
	})

	return g.Wait()
}
