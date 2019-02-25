package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/time/rate"

	ratelimitkit "github.com/go-kit/kit/ratelimit"
	"google.golang.org/grpc"

	"github.com/williamzion/vault"
	"github.com/williamzion/vault/pb"
)

func main() {
	var (
		httpAddr = flag.String("http", ":8080", "http listen address")
		gRPCAddr = flag.String("grpc", ":8081", "gRPC listen address")
	)
	flag.Parse()

	ctx := context.Background()
	srv := vault.NewService()
	errChan := make(chan error)

	// Traps termination signal (such as ctrl + C) and sends an error down
	// errChain.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	// Rate limit
	limit := rate.NewLimiter(rate.Every(1*time.Second), 1)

	hashEndpoint := vault.MakeHashEndpoint(srv)
	{
		hashEndpoint = ratelimitkit.NewDelayingLimiter(limit)(hashEndpoint)
	}
	validateEndpoint := vault.MakeValidateEndpoint(srv)
	{
		validateEndpoint = ratelimitkit.NewDelayingLimiter(limit)(validateEndpoint)
	}
	endpoints := vault.Endpoints{
		HashEndpoint:     hashEndpoint,
		ValidateEndpoint: validateEndpoint,
	}

	// HTTP transport
	go func() {
		log.Println("http:", *httpAddr)
		handler := vault.NewHTTPServer(ctx, endpoints)
		errChan <- http.ListenAndServe(*httpAddr, handler)
	}()

	// gRPC transport
	go func() {
		lis, err := net.Listen("tcp", *gRPCAddr)
		if err != nil {
			errChan <- err
			return
		}
		log.Println("grpc:", *gRPCAddr)
		handler := vault.NewGRPCServer(ctx, endpoints)
		gRPCServer := grpc.NewServer()
		pb.RegisterVaultServer(gRPCServer, handler)
		errChan <- gRPCServer.Serve(lis)
	}()

	log.Fatalln(<-errChan)
}
