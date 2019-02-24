package vault

import (
	"context"
	"testing"
)

func TestHasherService(t *testing.T) {
	srv := NewService()
	ctx := context.Background()
	h, err := srv.Hash(ctx, "password")
	if err != nil {
		t.Errorf("hash: %v", err)
	}
	ok, err := srv.Validate(ctx, "password", h)
	if err != nil {
		t.Errorf("validate: %v", err)
	}
	if !ok {
		t.Error("expected true from Validate")
	}

	ok, err := srv.Validate(ctx, "wrong password", h)
	if err != nil {
		t.Errorf("Validate: %v", err)
	}
	if ok {
		t.Error("expected false from Validate")
	}
}
