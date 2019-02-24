package vault

import (
	"context"
	"testing"
	"vault/pb"
)

func TestHasherService(t *testing.T) {
	svr := NewServer()
	ctx := context.Background()

	hr := pb.HashRequest{Password: "password"}
	h, err := svr.Hash(ctx, &hr)
	if err != nil {
		t.Errorf("hash: %v", err)
	}

	response, err := svr.Validate(ctx, &pb.ValidateRequest{Password: "password", Hash: h.Hash})
	if err != nil {
		t.Errorf("validate: %v", err)
	}
	if !response.Valid {
		t.Error("expected true from Valid")
	}

	response, err = svr.Validate(ctx, &pb.ValidateRequest{Password: "wrong password", Hash: h.Hash})
	if err != nil {
		t.Errorf("validate: %v", err)
	}
	if !response.Valid {
		t.Error("expected false from Valid")
	}
}
