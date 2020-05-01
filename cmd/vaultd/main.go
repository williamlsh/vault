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
	lightstep "github.com/lightstep/lightstep-tracer-go"
	opentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	zipkin "github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"sourcegraph.com/sourcegraph/appdash"
	appdashot "sourcegraph.com/sourcegraph/appdash/opentracing"

	"github.com/williamlsh/vault/internal/store"
	"github.com/williamlsh/vault/internal/vaultendpoint"
	"github.com/williamlsh/vault/internal/vaultransport"
	"github.com/williamlsh/vault/internal/vaultservice"
	vaultpb "github.com/williamlsh/vault/pb"
)

const vaultdLogLevel = "VAULTD_LOG_LEVEL"

func main() {
	var (
		httpAddr = flag.String("http-addr", ":443", "HTTP listen address")
		grpcAddr = flag.String("grpc-addr", ":8080", "gRPC listen address")
		promAddr = flag.String("prom-addr", ":8081", "Prometheus server listen address")
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
		// Zipkin tracer.
		zipkinURL = flag.String("zipkin-url", "", "Enable Zipkin tracing (zipkin-go-opentracing) using a reporter URL e.g. http://localhost:9411/api/v1/spans")
		// Lightstep tracer.
		lightstepToken = flag.String("lightstep-token", "", "Enable LightStep tracing via a LightStep access token")
		// Appdash.
		appdashAddr = flag.String("appdash-addr", "", "Enable Appdash tracing via an Appdash server host:port")
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

	// Tracer domian.
	var zipkinTracer *zipkin.Tracer
	{
		var (
			serviceName   = "vault-daemon"
			hostPort      = "localhost:8083"
			useNoopTracer = *zipkinURL == ""
			reporter      = zipkinhttp.NewReporter(*zipkinURL)
		)
		defer reporter.Close()

		endpoint, err := zipkin.NewEndpoint(serviceName, hostPort)
		if err != nil {
			level.Error(logger).Log("msg", "unable to create local endpoint", "err", err)
			os.Exit(1)
		}

		zipkinTracer, err = zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint), zipkin.WithNoopTracer(useNoopTracer))
		if err != nil {
			level.Error(logger).Log("msg", "unable to create tracer", "err", err)
			os.Exit(1)
		}
		if !useNoopTracer {
			level.Info(logger).Log("tracer", "Zipkin", "type", "Native", "URL", *zipkinURL)
		}
	}

	var tracer opentracing.Tracer
	{
		switch {
		case *zipkinURL != "":
			level.Info(logger).Log("tracer", "Zipkin", "type", "OpenTracing", "URL", *zipkinURL)
			tracer = zipkinot.Wrap(zipkinTracer)
			fallthrough
		case *lightstepToken != "":
			level.Info(logger).Log("tracer", "LightStep")
			tracer = lightstep.NewTracer(lightstep.Options{
				AccessToken: *lightstepToken,
			})
			fallthrough
		case *appdashAddr != "":
			level.Info(logger).Log("tracer", "Appdash", "addr", *appdashAddr)
			tracer = appdashot.NewTracer(appdash.NewRemoteCollector(*appdashAddr))
		default:
			tracer = opentracing.GlobalTracer() // no-op
		}
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
	var ints metrics.Counter
	{
		// Business-level metrics.
		ints = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "vault",
			Subsystem: "vaultsvc",
			Name:      "http_requests_summed",
			Help:      "Total count of http requests summed of all endpoints.",
		}, []string{})
	}

	// Datastore domain
	datastore := store.New(log.With(logger, "domain", "store"), dsn)

	// Service domain.
	var (
		service     = vaultservice.New(log.With(logger, "domain", "vaultservice"), ints, datastore)
		endpoints   = vaultendpoint.New(service, duration, tracer, zipkinTracer, log.With(logger, "domain", "vaultendpoint"))
		httpHandler = vaultransport.NewHTTPHandler(endpoints, tracer, zipkinTracer, log.With(logger, "domain", "vaultransport-http"))
		grpcServer  = vaultransport.NewGRPCServer(endpoints, tracer, zipkinTracer, log.With(logger, "domain", "vaultransport-grpc"))
	)

	errs := make(chan error, 2)

	// Metrics server.
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		errs <- http.ListenAndServeTLS(*promAddr, *tlsCert, *tlsKey, promhttp.Handler())
	}()

	// Interruption handler.
	go func() {
		c := make(chan os.Signal, 3)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
		errs <- fmt.Errorf("%s", <-c)
	}()

	// HTTP server with TLS.
	go func() {
		level.Info(logger).Log("transport", "HTTP", "addr", *httpAddr)
		errs <- http.ListenAndServeTLS(*httpAddr, *tlsCert, *tlsKey, httpHandler)
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
