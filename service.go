package vault

import (
	"context"

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
