package vault

import (
	"context"
	"encoding/json"
	"net/http"

	"vault/pb"

	"golang.org/x/crypto/bcrypt"
)

// server implements VaultServer interface.
type server struct{}

func (server) Hash(ctx context.Context, hr *pb.HashRequest) (*pb.HashResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(hr.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &pb.HashResponse{Hash: string(hash)}, nil
}

func (server) Validate(ctx context.Context, vr *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	err := bcrypt.CompareHashAndPassword([]byte(vr.Hash), []byte(vr.Password))
	if err != nil {
		return nil, err
	}
	return &pb.ValidateResponse{Valid: true}, nil
}

// NewServer returns a new VaultServer.
func NewServer() pb.VaultServer {
	return server{}
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
