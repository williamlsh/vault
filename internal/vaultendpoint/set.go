package vaultendpoint

import (
	"context"
	"time"

	stdjwt "github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"

	"github.com/williamlsh/vault/internal/vaultservice"
)

// SigningKey is a JWT signing key.
var SigningKey = []byte("zmh298onj30")

// Set collects all of the endpoints that compose a vault service.
type Set struct {
	HashEndpoint     endpoint.Endpoint
	ValidateEndpoint endpoint.Endpoint
}

// New returns a Set that wraps the provided server, and wires in all of the
// expected endpoint middlewares via the various parameters
func New(svc vaultservice.Service, duration metrics.Histogram, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) Set {
	jwtParser := jwt.NewParser(
		func(token *stdjwt.Token) (interface{}, error) { return SigningKey, nil }, stdjwt.SigningMethodHS256,
		jwt.StandardClaimsFactory,
	)

	var hashEndpoint endpoint.Endpoint
	{
		hashEndpoint = MakeHashEndpoint(svc)
		hashEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(hashEndpoint)
		hashEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(hashEndpoint)
		hashEndpoint = jwtParser(hashEndpoint)
		hashEndpoint = opentracing.TraceServer(otTracer, "Hash")(hashEndpoint)
		hashEndpoint = zipkin.TraceEndpoint(zipkinTracer, "Hash")(hashEndpoint)
		hashEndpoint = LoggingMiddleware(log.With(logger, "method", "Hash"))(hashEndpoint)
		hashEndpoint = InstrumentingMiddleware(duration.With("method", "Hash"))(hashEndpoint)
	}
	var validateEndpoint endpoint.Endpoint
	{
		validateEndpoint = MakeValidateEndpoint(svc)
		validateEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(validateEndpoint)
		validateEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(validateEndpoint)
		validateEndpoint = jwtParser(validateEndpoint)
		validateEndpoint = opentracing.TraceServer(otTracer, "Validate")(validateEndpoint)
		validateEndpoint = zipkin.TraceEndpoint(zipkinTracer, "Validate")(validateEndpoint)
		validateEndpoint = LoggingMiddleware(log.With(logger, "method", "Validate"))(validateEndpoint)
		validateEndpoint = InstrumentingMiddleware(duration.With("method", "Validate"))(validateEndpoint)
	}
	return Set{
		HashEndpoint:     hashEndpoint,
		ValidateEndpoint: validateEndpoint,
	}
}

// Hash implements vaultservice.Service interface, so Set may be used as a
// service. This is primarily  useful in the context of a client library.
func (s Set) Hash(ctx context.Context, password string) (string, error) {
	resp, err := s.HashEndpoint(ctx, HashRequest{Password: password})
	if err != nil {
		return "", err
	}
	response := resp.(HashResponse)
	return response.Hash, response.Err
}

// Validate implements vaultservice.Service interface, so Set may be used as a
// service. This is primarily  useful in the context of a client library.
func (s Set) Validate(ctx context.Context, password, hash string) (bool, error) {
	resp, err := s.ValidateEndpoint(ctx, ValidateRequest{Password: password, Hash: hash})
	if err != nil {
		return false, err
	}
	response := resp.(ValidateResponse)
	return response.Valid, response.Err
}

// MakeHashEndpoint constructs a Hash endpoint wrapping the service.
func MakeHashEndpoint(s vaultservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(HashRequest)
		v, err := s.Hash(ctx, req.Password)
		return HashResponse{Hash: v, Err: err}, nil
	}
}

// MakeValidateEndpoint constructs a Validate endpoint wrapping the service.
func MakeValidateEndpoint(s vaultservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ValidateRequest)
		v, err := s.Validate(ctx, req.Password, req.Hash)
		return ValidateResponse{Valid: v, Err: err}, nil
	}
}

// Compile time assertions for the response types implementing endpoint.Failer.
var (
	_ endpoint.Failer = HashResponse{}
	_ endpoint.Failer = ValidateResponse{}
)

type HashRequest struct {
	Password string `json:"password"`
}

type HashResponse struct {
	Hash string `json:"hash"`
	Err  error  `json:"-"`
}

func (r HashResponse) Failed() error {
	return r.Err
}

type ValidateRequest struct {
	Password string `json:"password"`
	Hash     string `json:"hash"`
}

type ValidateResponse struct {
	Valid bool  `json:"valid"`
	Err   error `json:"-"`
}

func (r ValidateResponse) Failed() error {
	return r.Err
}
