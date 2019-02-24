package vault

import (
	"context"
	"testing"
	"vault/pb"
)

func TestHasherService(t *testing.T) {
	srv := NewService()
	ctx := context.Background()

	hr := pb.HashRequest{Password: "password"}
	h, err := srv.Hash(ctx, &hr)
	if err != nil {
		t.Errorf("hash: %v", err)
	}

	resp, err := srv.Validate(ctx, &pb.ValidateRequest{Password: "password", Hash: h.Hash})
	if err != nil {
		t.Errorf("validate: %v", err)
	}
	if !resp.Valid {
		t.Error("expected true from Valid")
	}

	resp, err = srv.Validate(ctx, &pb.ValidateRequest{Password: "wrong password", Hash: h.Hash})
	if err == nil {
		t.Errorf("validate: %v", err)
	}
	t.SkipNow()
	if resp.Valid {
		t.Error("expected false from Valid")
	}
}
