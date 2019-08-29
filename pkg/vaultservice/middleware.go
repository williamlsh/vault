package vaultservice

import (
	"context"

	"github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/log"
)

// Middleware represents a service middleware.
type Middleware func(Service) Service

// LoggingMiddleware takes a logger as a dependency and returns a
// ServiceMiddleware.
func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return loggingMiddleware{
			logger: logger,
			next:   next,
		}
	}
}

type loggingMiddleware struct {
	logger log.Logger
	next   Service
}

func (mw loggingMiddleware) Hash(ctx context.Context, password string) (hash string, err error) {
	defer func() {
		mw.logger.Log("method", "Hash", "password", password, "hash", hash, "token", ctx.Value(jwt.JWTTokenContextKey).(string), "err", err)
	}()
	return mw.next.Hash(ctx, password)
}

func (mw loggingMiddleware) Validate(ctx context.Context, password, hash string) (v bool, err error) {
	defer func() {
		mw.logger.Log("method", "Validate", "password", password, "hash", hash, "valid", v, "token", ctx.Value(jwt.JWTTokenContextKey).(string), "err", err)
	}()
	return mw.next.Validate(ctx, password, hash)
}
