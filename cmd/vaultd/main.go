package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	vaultpb "github.com/williamzion/vault/pb"
	"github.com/williamzion/vault/pkg/vaultendpoint"
	"github.com/williamzion/vault/pkg/vaultransport"
	"github.com/williamzion/vault/pkg/vaultservice"
	"google.golang.org/grpc"
)

const vaultLogLevel = "VAULT_LOG_LEVEL"

func main() {
	fs := flag.NewFlagSet("vault", flag.ExitOnError)
	var (
		httpAddr = fs.String("http-addr", ":8080", "HTTP listen address")
		grpcAddr = fs.String("grpc-addr", ":8081", "gRPC listen address")
	)
	fs.Usage = usageFor(fs, os.Args[0]+" [flags]")
	fs.Parse(os.Args[1:])

	// Logger domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		// Note: Enable error level log in production mode.
		switch os.Getenv(vaultLogLevel) {
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

	// Service domian.
	var (
		service     = vaultservice.NewService()
		endpoints   = vaultendpoint.New(service, logger)
		httpHandler = vaultransport.NewHTTPHandler(endpoints, logger)
		grpcServer  = vaultransport.NewGRPCServer(endpoints, logger)
	)

	errs := make(chan error, 2)
	go func() {
		c := make(chan os.Signal, 3)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errs <- http.ListenAndServe(*httpAddr, httpHandler)
	}()

	go func() {
		lis, err := net.Listen("tcp", *grpcAddr)
		if err != nil {
			logger.Log("transport", "gRPC", "during", "Listen", "err", err)
			errs <- err
			return
		}
		logger.Log("transport", "gRPC", "addr", *grpcAddr)
		s := grpc.NewServer()
		vaultpb.RegisterVaultServer(s, grpcServer)
		errs <- s.Serve(lis)
	}()

	level.Error(logger).Log("exit", <-errs)
}

func usageFor(fs *flag.FlagSet, short string) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, " %s\n", short)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "FLAGS\n")
		w := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(w, "\t-%s %s\t%s\n", f.Name, f.DefValue, f.Usage)
		})
		w.Flush()
		fmt.Fprintf(os.Stderr, "\n")
	}
}
