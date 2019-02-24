package vault

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/endpoint"

	"golang.org/x/crypto/bcrypt"
)

// Service provides password hashing capacities.
type Service interface {
	Hash(ctx context.Context, password string) (string, error)
	Validate(ctx context.Context, password, hash string) (bool, error)
}

type vaultService struct{}

// NewService makes a new service.
func NewService() Service {
	return vaultService{}
}

func (vaultService) Hash(ctx context.Context, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (vaultService) Validate(ctx context.Context, password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(password, password, hash)
	if err != nil {
		return false, nil
	}
	return string(hash), nil
}

type hashRequest struct {
	Password string `json:"password"`
}

type hashResponse struct {
	Hash string `json:"hash"`
	Err  error  `json:"err,omitempty"`
}

// decodeHashRequest is a helper function that will decode the JSON body of
// http.Request to service.go.
// The signature for decodeHashRequest is dictated by Go kit because it will later use it to decode HTTP requests on our behalf.
func decodeHashRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req hashRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

type validateRequest struct {
	Password string `json:"password"`
	Hash     string `json:"hash"`
}

type validateResponse struct {
	Valid bool  `json:"valid"`
	Err   error `json:"err,omitempty"`
}

func decodeValidateRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req validateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

// MakeHashEndpoint turns Hash method  of vaultService into an endpoint.
func MakeHashEndpoint(srv Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(hashRequest)
		v, err := srv.Hash(ctx, req.Password)
		if err != nil {
			return hashResponse{v, err.Error()}, nil
		}
		return hashResponse{v, ""}, nil
	}
}

// MakeValidateEndpoint turns Validate method  of vaultService into an endpoint.
func MakeValidateEndpoint(srv Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(validateRequest)
		v, err := srv.Hash(ctx, req.Password, req.Hash)
		if err != nil {
			return validateResponse{false, err.Error()}, nil
		}
		return validateResponse{v, ""}, nil
	}
}

// Endpoints represents all endpoints for the vaultService.
type Endpoints struct {
	HashEndpoint, ValidEndpoint endpoint.Endpoint
}

// Hash uses the HashEndpoint to hash a password.
func (e Endpoints) Hash(ctx context.Context, password string) (string, error) {
	req := hashRequest{Password: password}
	resp, err := e.HashEndpoint(ctx, req)
	if err != nil {
		return "", err
	}
	hashResp := resp.(hashResponse)
	if err := hashResp.Err(); err != "" {
		return "", err
	}
	return hashResp.Hash, nil
}

// Validate uses the ValidEndpoint to validate a password a hash pair.
func (e Endpoints) Validate(ctx context.Context, password, hash string) (string, error) {
	req := validateRequest{Password: password, Hash: hash}
	resp, err := e.ValidEndpoint(ctx, req)
	if err != nil {
		return false, err
	}
	validateResp := resp.(validateResponse)
	if err := validateResp.Err(); err != "" {
		return "", err
	}
	return validateResp.Valid, nil
}
