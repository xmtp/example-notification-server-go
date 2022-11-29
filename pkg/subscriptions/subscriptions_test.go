package subscriptions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	database "github.com/xmtp/example-notification-server-go/pkg/db"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/logging"
	"github.com/xmtp/example-notification-server-go/test"
)

const INSTALLATION_ID = "installation_1"
const TOPIC = "topic1"

func createService(db *bun.DB) interfaces.Subscriptions {
	return NewSubscriptionsService(
		logging.CreateLogger("console", "info"),
		db,
	)
}

func Test_Subscribe(t *testing.T) {
	ctx := context.Background()
	db, cleanup := test.CreateTestDb()
	defer cleanup()

	svc := createService(db)

	err := svc.Subscribe(ctx, INSTALLATION_ID, []string{TOPIC})
	require.NoError(t, err)

	var stored database.Subscription
	err = db.NewSelect().
		Model(&stored).
		Where("installation_id = ?", INSTALLATION_ID).
		Scan(ctx)
	require.NoError(t, err)

	require.Equal(t, stored.InstallationId, INSTALLATION_ID)
	require.Equal(t, stored.IsActive, true)
	require.Equal(t, stored.Topic, TOPIC)
}

func Test_SubscribeMultiple(t *testing.T) {
	ctx := context.Background()
	db, cleanup := test.CreateTestDb()
	defer cleanup()

	svc := createService(db)

	topics := []string{"topic_1", "topic_2", "topic_3"}

	err := svc.Subscribe(ctx, INSTALLATION_ID, topics)
	require.NoError(t, err)

	stored := make([]database.Subscription, 0)
	err = db.NewSelect().
		Model(&stored).
		Where("installation_id = ?", INSTALLATION_ID).
		Scan(ctx)
	require.NoError(t, err)
	require.Len(t, stored, 3)

	for _, result := range stored {
		var found bool
		for _, topic := range topics {
			if result.Topic == topic {
				found = true
			}
		}
		require.True(t, found)
		require.Equal(t, result.InstallationId, INSTALLATION_ID)
		require.NotNil(t, result.CreatedAt)
	}
}

func Test_Unsubscribe(t *testing.T) {
	ctx := context.Background()
	db, cleanup := test.CreateTestDb()
	defer cleanup()

	svc := createService(db)

	err := svc.Subscribe(ctx, INSTALLATION_ID, []string{TOPIC})
	require.NoError(t, err)

	err = svc.Unsubscribe(ctx, INSTALLATION_ID, []string{TOPIC})
	require.NoError(t, err)

	var stored database.Subscription
	err = db.NewSelect().
		Model(&stored).
		Where("installation_id = ?", INSTALLATION_ID).
		Scan(ctx)
	require.NoError(t, err)
	require.Equal(t, stored.InstallationId, INSTALLATION_ID)
	require.False(t, stored.IsActive)
}

func Test_UnsubscribeResubscribe(t *testing.T) {
	ctx := context.Background()
	db, cleanup := test.CreateTestDb()
	defer cleanup()

	svc := createService(db)

	// Subscribe
	err := svc.Subscribe(ctx, INSTALLATION_ID, []string{TOPIC})
	require.NoError(t, err)

	// Unsubscribe
	err = svc.Unsubscribe(ctx, INSTALLATION_ID, []string{TOPIC})
	require.NoError(t, err)

	// Resubscribe
	err = svc.Subscribe(ctx, INSTALLATION_ID, []string{TOPIC})
	require.NoError(t, err)

	allResults := make([]database.Subscription, 0)
	err = db.NewSelect().
		Model(&allResults).
		Where("installation_id = ?", INSTALLATION_ID).
		Scan(ctx)
	require.NoError(t, err)
	require.Len(t, allResults, 1)

	stored := allResults[0]
	require.Equal(t, stored.InstallationId, INSTALLATION_ID)
	require.True(t, stored.IsActive)
}

func Test_GetSubscriptions(t *testing.T) {
	ctx := context.Background()
	db, cleanup := test.CreateTestDb()
	defer cleanup()

	svc := createService(db)

	err := svc.Subscribe(ctx, INSTALLATION_ID, []string{TOPIC})
	require.NoError(t, err)

	subs, err := svc.GetSubscriptions(ctx, TOPIC)
	require.NoError(t, err)
	require.Len(t, subs, 1)
}
