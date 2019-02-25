package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
	"vault"

	grpcclient "github.com/williamzion/vault/client/grpc"

	"google.golang.org/grpc"
)

func main() {
	var grpcAddr = flag.String("addr", ":8081", "grpc address")
	flag.Parse()

	ctx := context.Background()
	conn, err := grpc.Dial(*grpcAddr, grpc.WithInsecure(), grpc.WithTimeout(1*time.Second))
	if err != nil {
		log.Fatalln("gRPC dial:", err)
	}
	defer conn.Close()
	vaultService := grpcclient.New(conn)
	args := flag.Args()
	var cmd string
	cmd, args = pop(args)
	switch cmd {
	case "hash":
		var password string
		password, args = pop(args)
		hash(ctx, vaultService, password)
	case "validate":
		var password, hash string
		hash, args = pop(args)
		password, args = pop(args)
		validate(ctx, vaultService, password, hash)
	default:
		log.Fatalln("unknown command", cmd)
	}
}

func hash(ctx context.Context, service vault.Service, password string) {
	h, err := service.Hash(ctx, password)
	if err != nil {
		log.Fatalln(err.Error())
	}
	fmt.Println(h)
}

func validate(ctx context.Context, service vault.Service, password, hash string) {
	v, err := service.Validate(ctx, password, hash)
	if err != nil {
		log.Fatalln(err.Error())
	}
	if !v {
		fmt.Println("invalid:", v)
		os.Exit(1)
	}
	fmt.Println("valid:", v)
}

func pop(s []string) (string, []string) {
	if len(s) == 0 {
		return "", s
	}
	return s[0], s[1:]
}
