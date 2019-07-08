package vaultservice

import (
	"context"

	"github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/williamzion/vault/pkg/store"
	"golang.org/x/crypto/bcrypt"
)

// Service describes a service that hashes and validates passwords.
type Service interface {
	Hash(ctx context.Context, password string) (string, error)
	Validate(ctx context.Context, password, hash string) (bool, error)
}

type vaultService struct {
	logger log.Logger
	store  store.Store
}

// NewService makes a new service.
func NewService(logger log.Logger, s store.Store) Service {
	return &vaultService{
		logger: logger,
		store:  s,
	}
}

func (s *vaultService) Hash(ctx context.Context, password string) (string, error) {
	level.Info(s.logger).Log("during", "hash", "jwt_token", ctx.Value(jwt.JWTTokenContextKey).(string))

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		level.Error(s.logger).Log("during", "hash", "err", err)
		return "", err
	}
	errc := s.store.KeepSecret(hash)
	if err := <-errc; err != nil {
		level.Error(s.logger).Log("during", "keepSecret", "err", err)
		return "", err
	}
	level.Info(s.logger).Log("hash", "success")
	return string(hash), nil
}

func (s *vaultService) Validate(ctx context.Context, password, hash string) (bool, error) {
	level.Info(s.logger).Log("during", "validate", "jwt_token", ctx.Value(jwt.JWTTokenContextKey).(string))

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		level.Error(s.logger).Log("during", "validate", "err", err)
		return false, nil
	}
	level.Info(s.logger).Log("validate", "success")
	return true, nil
}
