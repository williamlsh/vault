package store

import (
	"github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
)

// Store represents a database store.
type Store interface {
	// KeepSecret keeps the password hash bytes in database.
	KeepSecret(secret []byte) <-chan error
}

// store implements Store interface.
type store struct {
	logger log.Logger
	db     *sqlx.DB
}

// New returns a new store.
func New(logger log.Logger, dsn string) Store {
	db := newDB(logger, dsn)
	return store{
		logger: logger,
		db:     db,
	}
}
