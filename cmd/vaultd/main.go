package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	vaultpb "github.com/williamlsh/vault/pb"
	"github.com/williamlsh/vault/pkg/store"
	"github.com/williamlsh/vault/pkg/vaultendpoint"
	"github.com/williamlsh/vault/pkg/vaultransport"
	"github.com/williamlsh/vault/pkg/vaultservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const vaultdLogLevel = "VAULTD_LOG_LEVEL"

func main() {
	var (
		httpAddr  = flag.String("http-addr", ":8080", "HTTP listen address")
		grpcAddr  = flag.String("grpc-addr", ":8081", "gRPC listen address")
		debugAddr = flag.String("debug-addr", ":8082", "Debug and metrics listen address")
		// TLS files.
		tlsCert = flag.String("tls-cert", "", "TLS certificate file")
		tlsKey  = flag.String("tls-key", "", "TLS key file")
		// Postgres connection credentials.
		pgUser    = flag.String("pg-user", "", "postgreSQL database username")
		pgPass    = flag.String("pg-password", "", "postgreSQL database password for provided user")
		pgDbname  = flag.String("pg-dbname", "", "postgreSQL database name")
		pgHost    = flag.String("pg-host", "localhost", "postgreSQL database host")
		pgSslmode = flag.String("pg-sslmode", "disable", "postgreSQL database connection sslmode option")
		pgPort    = flag.String("pg-port", "5432", "postgreSQL connection binding port")
	)
	flag.Parse()

	dsn := fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%s sslmode=%s", *pgUser, *pgDbname, *pgPass, *pgHost, *pgPort, *pgSslmode)

	// Logger domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		// Note: Enable error level log in production mode.
		switch os.Getenv(vaultdLogLevel) {
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

	// Metrics domain.
	var duration metrics.Histogram
	{
		// Endpoint-level metrics.
		duration = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "vault",
			Subsystem: "vaultsvc",
			Name:      "request_duration_seconds",
			Help:      "Request duration in seconds.",
		}, []string{"method", "success"})
	}

	// Datastore domain
	datastore := store.New(log.With(logger, "domain", "store"), dsn)

	// Service domain.
	var (
		service     = vaultservice.New(log.With(logger, "domain", "vaultservice"), datastore)
		endpoints   = vaultendpoint.New(service, log.With(logger, "domain", "vaultendpoint"), duration)
		httpHandler = vaultransport.NewHTTPHandler(endpoints, log.With(logger, "domain", "vaultransport-http"))
		grpcServer  = vaultransport.NewGRPCServer(endpoints, log.With(logger, "domain", "vaultransport-grpc"))
	)

	errs := make(chan error, 2)

	// Metrics server.
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		errs <- http.ListenAndServe(*debugAddr, nil)
	}()

	// Interruption handler.
	go func() {
		c := make(chan os.Signal, 3)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
		errs <- fmt.Errorf("%s", <-c)
	}()

	// HTTP server.
	go func() {
		level.Info(logger).Log("transport", "HTTP", "addr", *httpAddr)
		errs <- http.ListenAndServe(*httpAddr, httpHandler)
	}()

	// gRPC server.
	go func() {
		lis, err := net.Listen("tcp", *grpcAddr)
		if err != nil {
			level.Error(logger).Log("transport", "gRPC", "during", "Listen", "err", err)
			errs <- err
			return
		}
		level.Info(logger).Log("transport", "gRPC", "addr", *grpcAddr)
		// Create tls based credential.
		creds, err := credentials.NewServerTLSFromFile(*tlsCert, *tlsKey)
		if err != nil {
			level.Error(logger).Log("transport", "gRPC", "during", "construct TLS credentials", "err", err)
			errs <- err
			return
		}
		s := grpc.NewServer(grpc.Creds(creds))
		vaultpb.RegisterVaultServer(s, grpcServer)
		errs <- s.Serve(lis)
	}()

	// Waiting for error to be received.
	level.Error(logger).Log("exit", <-errs)
}
