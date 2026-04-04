package migrations

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var migrationFS embed.FS

// Legacy Bun deployments are considered equivalent to the first two golang-migrate
// migrations:
//   1. init schema
//   2. add subscription_hmac_keys + is_silent + unique subscription index
//
// Reconciliation must stay pinned to that handoff point so future golang-migrate-only
// migrations are still applied normally after older deployments upgrade.
const legacyBunBaselineVersion = 2

// LatestVersion returns the highest migration version found in the embedded
// migration files. This is useful for tests that need to assert the current
// schema version without hardcoding a number.
func LatestVersion() (int, error) {
	entries, err := migrationFS.ReadDir(".")
	if err != nil {
		return 0, err
	}
	max := 0
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || len(name) < 5 {
			continue
		}
		seq, err := strconv.Atoi(name[:5])
		if err != nil {
			continue
		}
		if seq > max {
			max = seq
		}
	}
	return max, nil
}

// Migrate always uses golang-migrate's schema_migrations table as the source of truth.
// For databases that were previously bootstrapped by Bun, we first "reconcile" by
// recording the fixed Bun-equivalent golang-migrate version in schema_migrations
// without replaying those baseline migrations. That handoff lets already-initialized
// deployments keep their existing application tables while still allowing future
// golang-migrate-only migrations to run normally after upgrade.
func Migrate(ctx context.Context, db *sql.DB) error {
	if err := reconcileExistingBunSchema(ctx, db); err != nil {
		return err
	}

	sourceDriver, err := iofs.New(migrationFS, ".")
	if err != nil {
		return err
	}
	defer func() {
		_ = sourceDriver.Close()
	}()

	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close()
	}()

	driver, err := postgres.WithConnection(ctx, conn, &postgres.Config{})
	if err != nil {
		return err
	}
	defer func() {
		_ = driver.Close()
	}()

	migrator, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = migrator.Close()
	}()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

// reconcileExistingBunSchema bridges databases that already have the legacy Bun-era
// application schema but do not yet have golang-migrate bookkeeping.
//
// Before reconciliation:
//   - the application tables already exist (installations, device_delivery_mechanisms,
//     subscriptions, subscription_hmac_keys, plus the legacy indexes/columns)
//   - schema_migrations is missing or empty
//
// After reconciliation:
//   - the application tables are unchanged
//   - schema_migrations exists and contains the fixed Bun-equivalent baseline version
//
// We intentionally do not translate Bun's bun_migrations metadata into golang-migrate
// rows. The application data tables are what matter for boot compatibility, so we detect
// the fully-initialized legacy schema directly and mark the new migration runner at the
// Bun handoff version rather than at whatever the latest embedded migration happens to be.
func reconcileExistingBunSchema(ctx context.Context, db *sql.DB) error {
	alreadyTracked, err := hasSchemaMigrationState(ctx, db)
	if err != nil {
		return err
	}
	if alreadyTracked {
		return nil
	}

	exists, err := hasLegacySchema(ctx, db)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version bigint NOT NULL PRIMARY KEY,
			dirty boolean NOT NULL
		)`); err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO schema_migrations (version, dirty)
		VALUES ($1, FALSE)
		ON CONFLICT (version) DO UPDATE SET dirty = EXCLUDED.dirty
	`, legacyBunBaselineVersion)
	return err
}

func hasSchemaMigrationState(ctx context.Context, db *sql.DB) (bool, error) {
	var tableExists bool
	if err := db.QueryRowContext(ctx, `SELECT to_regclass('public.schema_migrations') IS NOT NULL`).Scan(&tableExists); err != nil {
		return false, err
	}
	if !tableExists {
		return false, nil
	}

	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

func hasLegacySchema(ctx context.Context, db *sql.DB) (bool, error) {
	checks := []string{
		`SELECT to_regclass('public.installations') IS NOT NULL`,
		`SELECT to_regclass('public.device_delivery_mechanisms') IS NOT NULL`,
		`SELECT to_regclass('public.subscriptions') IS NOT NULL`,
		`SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = 'public'
			  AND table_name = 'subscriptions'
			  AND column_name = 'is_silent'
		)`,
		`SELECT to_regclass('public.subscriptions_installation_id_topic_idx') IS NOT NULL`,
		`SELECT to_regclass('public.subscription_hmac_keys') IS NOT NULL`,
	}

	for _, query := range checks {
		var ok bool
		if err := db.QueryRowContext(ctx, query).Scan(&ok); err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}

	return true, nil
}
