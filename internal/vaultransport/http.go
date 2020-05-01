package vaultransport

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	stdjwt "github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"

	"github.com/williamlsh/vault/internal/vaultendpoint"
	"github.com/williamlsh/vault/internal/vaultservice"
)

// NewHTTPHandler returns an HTTP handler thant makes a set of endpoints
// available on predefined paths.
func NewHTTPHandler(endpoints vaultendpoint.Set, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) http.Handler {
	options := []httptransport.ServerOption{
		httptransport.ServerBefore(jwt.HTTPToContext()),
		httptransport.ServerErrorEncoder(errorEncoder),
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		zipkin.HTTPServerTrace(zipkinTracer),
	}

	m := http.NewServeMux()
	m.Handle("/hash", httptransport.NewServer(
		endpoints.HashEndpoint,
		decodeHTTPHashRequest,
		encodeHTTPGenericResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "Hash", logger)))...,
	))
	m.Handle("/validate", httptransport.NewServer(
		endpoints.ValidateEndpoint,
		decodeHTTPValidateRequest,
		encodeHTTPGenericResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "Validate", logger)))...,
	))
	return m
}

// NewHTTPClient returns an VaultService backed by an HTTP server living at the
// remote instance. We expect instance to come from a service discovery system,
// so likely of the form "host:port". We bake-in certain middleware,
// implementing the client library pattern.
func NewHTTPClient(instance string, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) (vaultservice.Service, error) {
	if !strings.HasPrefix(instance, "https") {
		instance = "https://" + instance
	}
	u, err := url.Parse(instance)
	if err != nil {
		return nil, err
	}

	// Use customized http client, especially useful for localhost TLS test.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	options := []httptransport.ClientOption{
		httptransport.ClientBefore(opentracing.ContextToHTTP(otTracer, logger)),
		httptransport.ClientBefore(jwt.ContextToHTTP()),
		httptransport.SetClient(client),
		zipkin.HTTPClientTrace(zipkinTracer),
	}

	limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 100))

	// Client scope JWT signer.
	jwtSigner := jwt.NewSigner(
		kid,
		vaultendpoint.SigningKey,
		stdjwt.SigningMethodHS256,
		stdjwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(tokExp).Unix(),
		},
	)

	var hashEndpoint endpoint.Endpoint
	{
		hashEndpoint = httptransport.NewClient(
			"POST",
			copyURL(u, "/hash"),
			encodeHTTPGenericRequest,
			decodeHTTPHashResponse,
			options...,
		).Endpoint()
		hashEndpoint = opentracing.TraceClient(otTracer, "Hash")(hashEndpoint)
		hashEndpoint = zipkin.TraceEndpoint(zipkinTracer, "Hash")(hashEndpoint)
		hashEndpoint = jwtSigner(hashEndpoint)
		hashEndpoint = limiter(hashEndpoint)
		hashEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Hash",
			Timeout: 30 * time.Second,
		}))(hashEndpoint)
	}
	var validateEndpoint endpoint.Endpoint
	{
		validateEndpoint = httptransport.NewClient(
			"POST",
			copyURL(u, "/validate"),
			encodeHTTPGenericRequest,
			decodeHTTPValidateResponse,
			options...,
		).Endpoint()
		validateEndpoint = opentracing.TraceClient(otTracer, "Validate")(validateEndpoint)
		validateEndpoint = zipkin.TraceEndpoint(zipkinTracer, "Validate")(validateEndpoint)
		validateEndpoint = jwtSigner(validateEndpoint)
		validateEndpoint = limiter(validateEndpoint)
		validateEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Validate",
			Timeout: 10 * time.Second,
		}))(validateEndpoint)
	}
	return vaultendpoint.Set{
		HashEndpoint:     hashEndpoint,
		ValidateEndpoint: validateEndpoint,
	}, nil
}

func copyURL(base *url.URL, path string) *url.URL {
	next := *base
	next.Path = path
	return &next
}

func errorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	w.WriteHeader(err2code(err))
	json.NewEncoder(w).Encode(errorWrapper{Error: err.Error()})
}

func err2code(err error) int {
	// In practical use case, a error switch will be used.
	return http.StatusInternalServerError
}

func errDecoder(r *http.Response) error {
	var w errorWrapper
	if err := json.NewDecoder(r.Body).Decode(&w); err != nil {
		return err
	}
	return errors.New(w.Error)
}

type errorWrapper struct {
	Error string `json:"error"`
}

// decodeHTTPHashRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded hash request from the HTTP request body. Primarily useful in a
// server.
func decodeHTTPHashRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req vaultendpoint.HashRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func decodeHTTPValidateRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req vaultendpoint.ValidateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

// decodeHTTPHashResponse is a transport/http.DecodeResponseFunc that decodes a
// JSON-encoded hash response from the HTTP response body. If the response has
// a non-200 status code, we will interpret that as an error and attempt to
// decode the specific error message form the response body. Primarily useful
// in a client.
func decodeHTTPHashResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	var resp vaultendpoint.HashResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

func decodeHTTPValidateResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	var resp vaultendpoint.ValidateResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

// encodeHTTPGenericRequest is a transport/http.DecodeRequestFunc that
// JSON-encodes any request to the request body. Primarily useful in a client.
func encodeHTTPGenericRequest(_ context.Context, r *http.Request, request interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)
	return nil
}

func encodeHTTPGenericResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if f, ok := response.(endpoint.Failer); ok && f.Failed() != nil {
		errorEncoder(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
