package vault

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"vault/pb"

	"github.com/go-kit/kit/endpoint"
	"golang.org/x/crypto/bcrypt"
)

// service implements VaultServer interface.
type service struct{}

func (service) Hash(ctx context.Context, hr *pb.HashRequest) (*pb.HashResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(hr.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &pb.HashResponse{Hash: string(hash)}, nil
}

func (service) Validate(ctx context.Context, vr *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	err := bcrypt.CompareHashAndPassword([]byte(vr.Hash), []byte(vr.Password))
	if err != nil {
		return nil, err
	}
	return &pb.ValidateResponse{Valid: true}, nil
}

// NewService returns a new VaultServer.
func NewService() pb.VaultServer {
	return service{}
}

// decodeHashRequest is helper function dictated by Go kit to decode hash
// request for Hash. The original signature from Go kit is http.DecodeRequestFunc.
// See: https://github.com/go-kit/kit/blob/master/transport/grpc/encode_decode.go
func decodeHashRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var hr *pb.HashRequest
	err := json.NewDecoder(r.Body).Decode(hr)
	if err != nil {
		return nil, err
	}
	return hr, nil
}

// decodeValidateRequest is a helper function for Validate.
func decodeValidateRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var vr *pb.ValidateRequest
	err := json.NewDecoder(r.Body).Decode(vr)
	if err != nil {
		return nil, err
	}
	return vr, nil
}

// encodeResponse is a helper function for Go kit.
func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

// MakeHashEndpoint turns Hash to a Go kit Endpoint.
// Never return error in transport layer.
func MakeHashEndpoint(srv pb.VaultServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		r := request.(*pb.HashRequest)
		h, err := srv.Hash(ctx, r)
		if err != nil {
			return nil, nil
		}
		return h, nil
	}
}

// MakeValidateEndpoint turns Validate to Go kit Endpoint
func MakeValidateEndpoint(srv pb.VaultServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		r := request.(*pb.ValidateRequest)
		h, err := srv.Validate(ctx, r)
		if err != nil {
			return nil, nil
		}
		return h, nil
	}
}

// Endpoints represents all endpoints for VaultServer service.
type Endpoints struct {
	HashEndpoint, ValidateEndpoint endpoint.Endpoint
}

// Hash uses the HashEndpoint to hash a password.
func (e Endpoints) Hash(ctx context.Context, hr *pb.HashRequest) (*pb.HashResponse, error) {
	resp, err := e.HashEndpoint(ctx, hr)
	if err != nil {
		return nil, err
	}
	hashResp := resp.(*pb.HashResponse)
	if hashResp == nil {
		return nil, errors.New("hash request failed")
	}
	return hashResp, nil
}

// Validate uses the ValidateEndpoint to validate a password and a hash pair.
func (e Endpoints) Validate(ctx context.Context, hr *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	resp, err := e.ValidateEndpoint(ctx, hr)
	if err != nil {
		return nil, err
	}
	validateResp := resp.(*pb.ValidateResponse)
	if validateResp == nil {
		return nil, errors.New("validate request failed")
	}
	return validateResp, nil
}
