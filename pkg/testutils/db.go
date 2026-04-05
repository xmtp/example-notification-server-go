package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	database "github.com/xmtp/example-notification-server-go/pkg/db"
)

const TEST_DSN = "postgres://postgres:xmtp@localhost:25432/postgres?sslmode=disable"

var dbNameSanitizer = regexp.MustCompile(`[^a-z0-9_]+`)

func CreateTestDb(t *testing.T) *sql.DB {
	t.Helper()

	db := CreateEmptyTestDb(t)
	if err := database.Migrate(t.Context(), db); err != nil {
		t.Fatal(err)
	}
	return db
}

func CreateEmptyTestDb(t *testing.T) *sql.DB {
	t.Helper()

	dbName := uniqueDatabaseName(t.Name())
	adminDB := openDatabase(t, TEST_DSN)
	createDatabase(t, adminDB, dbName)
	_ = adminDB.Close()

	dsn := databaseDSN(t, dbName)
	db, err := database.CreateDB(dsn, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = db.Close()

		adminDB := openDatabase(t, TEST_DSN)
		defer func() {
			_ = adminDB.Close()
		}()

		_, _ = adminDB.ExecContext(
			context.Background(),
			"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()",
			dbName,
		)
		_, _ = adminDB.ExecContext(context.Background(), fmt.Sprintf(`DROP DATABASE IF EXISTS "%s"`, dbName))
	})

	return db
}

func openDatabase(t *testing.T, dsn string) *sql.DB {
	t.Helper()

	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	db := stdlib.OpenDB(*config)
	if err := db.PingContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	return db
}

func createDatabase(t *testing.T, adminDB *sql.DB, dbName string) {
	t.Helper()

	_, err := adminDB.ExecContext(context.Background(), fmt.Sprintf(`CREATE DATABASE "%s"`, dbName))
	if err != nil {
		t.Fatal(err)
	}
}

func databaseDSN(t *testing.T, dbName string) string {
	t.Helper()

	parsed, err := url.Parse(TEST_DSN)
	if err != nil {
		t.Fatal(err)
	}
	parsed.Path = "/" + dbName
	return parsed.String()
}

func uniqueDatabaseName(testName string) string {
	name := strings.ToLower(testName)
	name = strings.ReplaceAll(name, "/", "_")
	name = dbNameSanitizer.ReplaceAllString(name, "_")
	name = strings.Trim(name, "_")
	return fmt.Sprintf("test_%s_%d", name, time.Now().UnixNano())
}
