package vaultendpoint

import (
	"context"

	stdjwt "github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/williamzion/vault/pkg/vaultservice"
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
func New(svc vaultservice.Service, logger log.Logger) Set {
	parser := jwt.NewParser(
		func(token *stdjwt.Token) (interface{}, error) { return SigningKey, nil }, stdjwt.SigningMethodHS256,
		jwt.StandardClaimsFactory,
	)

	var hashEndpoint endpoint.Endpoint
	{
		hashEndpoint = MakeHashEndpoint(svc)
		hashEndpoint = parser(hashEndpoint)
	}
	var validateEndpoint endpoint.Endpoint
	{
		validateEndpoint = MakeValidateEndpoint(svc)
		validateEndpoint = parser(validateEndpoint)
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
