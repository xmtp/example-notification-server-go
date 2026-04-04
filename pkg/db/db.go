package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/xmtp/example-notification-server-go/pkg/db/migrations"
)

func CreateDB(dsn string, waitForDB time.Duration) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	waitUntil := time.Now().Add(waitForDB)
	err = db.Ping()
	for err != nil && time.Now().Before(waitUntil) {
		time.Sleep(3 * time.Second)
		err = db.Ping()
	}
	if err != nil {
		_ = db.Close()
		return nil, errors.New("timeout waiting for db")
	}

	return db, nil
}

func Migrate(ctx context.Context, db *sql.DB) error {
	return migrations.Migrate(ctx, db)
}

func CreateMigrationFiles(name string) ([]migrations.File, error) {
	return migrations.CreateFiles("pkg/db/migrations", name)
}

func LatestMigrationVersion() (int, error) {
	return migrations.LatestVersion()
}
