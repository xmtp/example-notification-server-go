package db_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	database "github.com/xmtp/example-notification-server-go/pkg/db"
	testdb "github.com/xmtp/example-notification-server-go/pkg/testutils"
)

func TestMigrateFreshDatabase(t *testing.T) {
	db := createRawDB(t)
	resetSchema(t, db)

	require.NoError(t, database.Migrate(t.Context(), db))
	require.NoError(t, database.Migrate(t.Context(), db))

	assertRelationExists(t, db, "installations")
	assertRelationExists(t, db, "device_delivery_mechanisms")
	assertRelationExists(t, db, "subscriptions")
	assertRelationExists(t, db, "subscription_hmac_keys")
	assertRelationExists(t, db, "subscriptions_installation_id_topic_idx")
	assertRelationExists(t, db, "device_delivery_mechanisms_latest_idx")
}

func TestMigrateExistingLegacySchema(t *testing.T) {
	db := createRawDB(t)
	resetSchema(t, db)

	legacyStatements := []string{
		`CREATE TABLE installations (
			id TEXT PRIMARY KEY,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE TABLE device_delivery_mechanisms (
			id BIGSERIAL PRIMARY KEY,
			installation_id TEXT NOT NULL REFERENCES installations(id),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			kind TEXT NOT NULL,
			token TEXT NOT NULL,
			UNIQUE (installation_id, kind, token)
		)`,
		`CREATE TABLE subscriptions (
			id BIGSERIAL PRIMARY KEY,
			installation_id TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			topic TEXT NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT FALSE,
			is_silent BOOLEAN NOT NULL DEFAULT FALSE
		)`,
		`CREATE INDEX subscriptions_topic_is_active_idx ON subscriptions (topic, is_active)`,
		`CREATE INDEX device_delivery_mechanisms_installation_id_idx ON device_delivery_mechanisms (installation_id)`,
		`CREATE UNIQUE INDEX subscriptions_installation_id_topic_idx ON subscriptions (installation_id, topic)`,
		`CREATE TABLE subscription_hmac_keys (
			subscription_id BIGINT NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
			thirty_day_periods_since_epoch INTEGER NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			key BYTEA NOT NULL,
			PRIMARY KEY (subscription_id, thirty_day_periods_since_epoch)
		)`,
	}
	for _, statement := range legacyStatements {
		_, err := db.ExecContext(t.Context(), statement)
		require.NoError(t, err)
	}

	require.NoError(t, database.Migrate(t.Context(), db))

	var version int
	err := db.QueryRowContext(t.Context(), `SELECT version FROM schema_migrations`).Scan(&version)
	require.NoError(t, err)

	latest, latestErr := database.LatestMigrationVersion()
	require.NoError(t, latestErr)
	require.Equal(t, latest, version)
}

func TestMigration_BinaryTopics_DataConversion(t *testing.T) {
	db := testdb.CreateEmptyTestDb(t)

	// Run migrations up through 00003 (db_perf_improvements, pre-binary schema)
	require.NoError(t, database.MigrateUpTo(t.Context(), db, 3))

	// Seed rows: conforming group, conforming welcome, non-conforming, duplicate (case diff)
	_, err := db.ExecContext(t.Context(), `
		INSERT INTO installations (id) VALUES ('inst1');
		INSERT INTO subscriptions (installation_id, topic, is_active) VALUES
			('inst1', '/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto', TRUE),
			('inst1', '/xmtp/mls/1/w-f3ac64eba2272334/proto', TRUE),
			('inst1', '/xmtp/mls/1/w-test_installation/proto', TRUE),
			('inst1', '/xmtp/mls/1/g-24CE39D660600B3A98ADFF3075B6D1F4/proto', TRUE);
	`)
	require.NoError(t, err)

	// Run migration 00004 (binary conversion + drop legacy column)
	require.NoError(t, database.MigrateUpTo(t.Context(), db, 4))

	// Verify: group topic converted correctly (first byte 0x00 = TopicKindGroupMessagesV1)
	var topicBytes []byte
	err = db.QueryRowContext(t.Context(),
		"SELECT topic FROM subscriptions WHERE installation_id = 'inst1' AND get_byte(topic, 0) = 0 LIMIT 1",
	).Scan(&topicBytes)
	require.NoError(t, err)
	require.Equal(t, byte(0x00), topicBytes[0]) // TopicKindGroupMessagesV1

	// Verify: non-conforming row deleted, duplicate collapsed — 2 rows remain
	var count int
	err = db.QueryRowContext(t.Context(),
		"SELECT COUNT(*) FROM subscriptions WHERE installation_id = 'inst1'",
	).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 2, count)

	// Verify topic_legacy column does not exist
	err = db.QueryRowContext(t.Context(),
		"SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'subscriptions' AND column_name = 'topic_legacy'",
	).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)
}

func createRawDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := database.CreateDB(testdb.TEST_DSN, 5*time.Second)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func resetSchema(t *testing.T, db *sql.DB) {
	t.Helper()

	statements := []string{
		"DROP TABLE IF EXISTS schema_migrations",
		"DROP TABLE IF EXISTS bun_migrations",
		"DROP TABLE IF EXISTS subscription_hmac_keys",
		"DROP TABLE IF EXISTS subscriptions",
		"DROP TABLE IF EXISTS device_delivery_mechanisms",
		"DROP TABLE IF EXISTS installations",
	}
	for _, statement := range statements {
		_, err := db.ExecContext(t.Context(), statement)
		require.NoError(t, err)
	}
}

func assertRelationExists(t *testing.T, db *sql.DB, name string) {
	t.Helper()

	var exists bool
	err := db.QueryRowContext(t.Context(), `SELECT to_regclass($1) IS NOT NULL`, "public."+name).Scan(&exists)
	require.NoError(t, err)
	require.True(t, exists, name)
}
