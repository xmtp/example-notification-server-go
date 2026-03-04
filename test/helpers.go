package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/extra/bundebug"
	database "github.com/xmtp/example-notification-server-go/pkg/db"
)

const TEST_DSN = "postgres://postgres:xmtp@localhost:25432/postgres?sslmode=disable"

func createDb(t *testing.T) *bun.DB {
	t.Helper()

	db, err := database.CreateBunDB(TEST_DSN, 5*time.Second)
	require.NoError(t, err)

	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	return db
}

func CreateTestDb(t *testing.T) (*bun.DB, func()) {
	t.Helper()

	ctx := context.Background()
	db := createDb(t)
	_ = database.Migrate(ctx, db)

	return db, func() {
		_, _ = db.NewTruncateTable().Model((*database.Installation)(nil)).Cascade().Exec(ctx)
		_, _ = db.NewTruncateTable().Model((*database.DeviceDeliveryMechanism)(nil)).Cascade().Exec(ctx)
		_, _ = db.NewTruncateTable().Model((*database.Subscription)(nil)).Cascade().Exec(ctx)
		_, _ = db.NewTruncateTable().Model((*database.SubscriptionHmacKeys)(nil)).Cascade().Exec(ctx)
	}
}
