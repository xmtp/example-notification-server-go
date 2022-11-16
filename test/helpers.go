package test

import (
	"context"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/extra/bundebug"
	database "github.com/xmtp/example-notification-server-go/pkg/db"
)

const TEST_DSN = "postgres://postgres:xmtp@localhost:25432/postgres?sslmode=disable"

func createDb() *bun.DB {
	db, _ := database.CreateBunDB(TEST_DSN, 5*time.Second)
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	return db
}

func CreateTestDb() (*bun.DB, func()) {
	ctx := context.Background()
	db := createDb()
	_ = database.Migrate(ctx, db)

	return db, func() {
		_ = db.ResetModel(ctx, (*database.Installation)(nil), (*database.Subscription)(nil), (*database.DeviceDeliveryMechanism)(nil))
	}
}
