package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/williamzion/vault/pkg/vaultransport"
	"github.com/williamzion/vault/pkg/vaultservice"
	"google.golang.org/grpc"
)

const (
	vaultcliLogLevel = "VAULTCLI_LOG_LEVEL"
	grpcDialTimeout  = 1 * time.Second
	rpcTimeout       = 2 * time.Second
)

func main() {
	var (
		httpAddr = flag.String("http-addr", ":8080", "HTTP listen address")
		grpcAddr = flag.String("grpc-addr", ":8081", "gRPC listen address")
		method   = flag.String("method", "", "hash, validate")
	)
	flag.Parse()

	// Logger domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		// Note: Enable error level log in production mode.
		switch os.Getenv(vaultcliLogLevel) {
		case "debug", "all":
			logger = level.NewFilter(logger, level.AllowAll())
		case "info":
			logger = level.NewFilter(logger, level.AllowInfo())
		case "warn":
			logger = level.NewFilter(logger, level.AllowWarn())
		case "error":
			logger = level.NewFilter(logger, level.AllowError())
		case "none":
			logger = level.NewFilter(logger, level.AllowNone())
		default:
			logger = level.NewFilter(logger, level.AllowError())
		}
		logger = log.With(logger, "ts", log.DefaultTimestamp)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var (
		svc vaultservice.Service
		err error
	)
	if *httpAddr != "" {
		svc, err = vaultransport.NewHTTPClient(*httpAddr, logger)
	} else if *grpcAddr != "" {
		conn, err := grpc.Dial(*grpcAddr, grpc.WithInsecure(), grpc.WithTimeout(grpcDialTimeout))
		if err != nil {
			level.Error(logger).Log("transport", "gRPC", "during", "grpc dial", "err", err)
			os.Exit(1)
		}
		defer conn.Close()
		svc = vaultransport.NewGRPCClient(conn, logger)
	} else {
		level.Error(logger).Log("err", "no remote address specified")
		os.Exit(1)
	}
	if err != nil {
		level.Error(logger).Log("err", err)
		// Note: don't use os.Exit(1) here and below, or deferred func will not
		// execute.
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), rpcTimeout)
	defer cancel()
	switch *method {
	case "hash":
		h, err := svc.Hash(ctx, "znm9832nmrfz4egwy43rn8")
		if err != nil {
			level.Error(logger).Log("method", "hash", "err", err)
			return
		}
		level.Info(logger).Log("method", "hash", "result", h)
	case "validate":
		v, err := svc.Validate(ctx, "znm9832nmrfz4egwy43rn8", "$2a$10$8e4JwCH9mCppJpTQ3Ax1PevFIt79her0oOg7AFy3eA4BNoeOMX1w.")
		if err != nil {
			level.Error(logger).Log("method", "validate", "err", err)
			return
		}
		level.Info(logger).Log("method", "validate", "result", v)
	default:
		level.Error(logger).Log("err", "invalid method")
		return
	}
}
