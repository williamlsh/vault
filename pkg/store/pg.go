package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgres driver
)

const (
	driverName = "postgres"
	sqlTimout  = 1 * time.Second
)

func newDB(logger log.Logger, dsn string) *sqlx.DB {
	db := sqlx.MustConnect(driverName, dsn)
	level.Info(logger).Log("database", driverName, "connect", "success")
	return db
}

// KeepSecret keeps the password hash bytes in database.
func (s store) KeepSecret(secret []byte) <-chan error {
	errc := make(chan error, 1)
	go keepSecret(s.logger, s.db, secret, errc)
	return errc
}

func keepSecret(logger log.Logger, db *sqlx.DB, secret []byte, errc chan<- error) {
	q := `insert into secret (hash) values ($1);`

	ctx, cancel := context.WithTimeout(context.Background(), sqlTimout)
	defer cancel()

	tx, err := db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	_, err = tx.ExecContext(ctx, q, secret)

	if err != nil {
		level.Error(logger).Log("during", "transaction exec", "err", err)
		tx.Rollback()
		errc <- err
		return
	}
	err = tx.Commit()
	if err != nil {
		level.Error(logger).Log("during", "transaction commit", "err", err)
		errc <- err
		return
	}

	level.Info(logger).Log("keepSecret", "success")
	errc <- nil
}
