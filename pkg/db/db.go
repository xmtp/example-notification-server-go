package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"log"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
	"github.com/xmtp/example-notification-server-go/pkg/db/migrations"
)

func CreateBunDB(dsn string, waitForDB time.Duration) (*bun.DB, error) {
	db, err := createDB(dsn, waitForDB)
	if err != nil {
		return nil, err
	}
	return bun.NewDB(db, pgdialect.New()), nil
}

func createDB(dsn string, waitForDB time.Duration) (*sql.DB, error) {
	db := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	waitUntil := time.Now().Add(waitForDB)
	err := db.Ping()
	for err != nil && time.Now().Before(waitUntil) {
		time.Sleep(3 * time.Second)
		err = db.Ping()
	}
	if err != nil {
		return nil, errors.New("timeout waiting for db")
	}
	return db, nil
}

func Migrate(ctx context.Context, db *bun.DB) error {
	migrator := migrate.NewMigrator(db, migrations.Migrations)
	err := migrator.Init(ctx)
	if err != nil {
		return err
	}

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return err
	}

	if group.IsZero() {
		log.Println("No new migrations to run")
	}

	return nil
}
