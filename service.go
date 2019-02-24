package vault

import (
	"context"
	"encoding/json"
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
func MakeHashEndpoint(srv pb.VaultServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		r := request.(*pb.HashRequest)
		h, err := srv.Hash(ctx, r)
		if err != nil {
			return nil, err
		}
		return h, nil
	}
}
