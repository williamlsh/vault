package vaultservice

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/williamlsh/vault/pkg/store"
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

// New makes a new service.
func New(logger log.Logger, s store.Store) Service {
	var svc Service
	{
		svc = newBasicService(logger, s)
		svc = LoggingMiddleware(logger)(svc)
	}
	return svc
}

func newBasicService(logger log.Logger, s store.Store) Service {
	return &vaultService{
		logger: logger,
		store:  s,
	}
}

func (s *vaultService) Hash(ctx context.Context, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	errc := s.store.KeepSecret(hash)
	if err := <-errc; err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *vaultService) Validate(ctx context.Context, password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return false, nil
	}
	return true, nil
}
