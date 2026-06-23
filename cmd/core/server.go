package main

import (
	"context"
	"log"
	"net"
	"os"

	ledgerpb "github.com/fair-n-square-co/apis/gen/pkg/fairnsquare/service/ledger/v1alpha1"
	"github.com/fair-n-square-co/core/internal/ledger/service"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func server() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	port := ":8080"
	envPort := os.Getenv("PORT")
	if envPort != "" {
		port = ":" + envPort
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("unable to listen on port ", port)
	}
	defer func(lis net.Listener) {
		err := lis.Close()
		if err != nil {
			log.Println(err)
		}
	}(lis)

	opts := []grpc.ServerOption{grpc.Creds(insecure.NewCredentials())}
	grpcServer := grpc.NewServer(opts...)
	reflection.Register(grpcServer)

	ledgerServer := service.NewFriendServer()
	ledgerpb.RegisterFriendServiceServer(grpcServer, ledgerServer)

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		log.Printf("listening on port %s", port)
		return grpcServer.Serve(lis)
	})

	if err := g.Wait(); err != nil {
		log.Fatalf("server error: %v", err)
		return err
	}

	return nil
}
