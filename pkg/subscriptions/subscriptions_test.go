package subscriptions

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/testutils"
)

const INSTALLATION_ID = "installation_1"
const TOPIC = "topic1"

type storedSubscription struct {
	ID             int64
	CreatedAt      time.Time
	InstallationID string
	Topic          string
	IsActive       bool
	IsSilent       bool
}

func createService(t *testing.T, db *sql.DB) interfaces.Subscriptions {
	t.Helper()
	return NewSubscriptionsService(
		testutils.TestLogger(t),
		db,
	)
}

func fetchSubscriptions(t *testing.T, ctx context.Context, db *sql.DB, installationID string) []storedSubscription {
	t.Helper()

	rows, err := db.QueryContext(
		ctx,
		`SELECT id, created_at, installation_id, topic, is_active, is_silent
		 FROM subscriptions
		 WHERE installation_id = $1
		 ORDER BY id ASC`,
		installationID,
	)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, rows.Close())
	}()

	var results []storedSubscription
	for rows.Next() {
		var row storedSubscription
		require.NoError(t, rows.Scan(
			&row.ID,
			&row.CreatedAt,
			&row.InstallationID,
			&row.Topic,
			&row.IsActive,
			&row.IsSilent,
		))
		results = append(results, row)
	}
	require.NoError(t, rows.Err())

	return results
}

func Test_Subscribe(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)

	svc := createService(t, db)

	err := svc.Subscribe(ctx, INSTALLATION_ID, []string{TOPIC})
	require.NoError(t, err)

	stored := fetchSubscriptions(t, ctx, db, INSTALLATION_ID)
	require.Len(t, stored, 1)
	require.Equal(t, INSTALLATION_ID, stored[0].InstallationID)
	require.True(t, stored[0].IsActive)
	require.Equal(t, TOPIC, stored[0].Topic)
}

func Test_SubscribeMultiple(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)
	svc := createService(t, db)

	topics := []string{"topic_1", "topic_2", "topic_3"}
	err := svc.Subscribe(ctx, INSTALLATION_ID, topics)
	require.NoError(t, err)

	stored := fetchSubscriptions(t, ctx, db, INSTALLATION_ID)
	require.Len(t, stored, 3)
	for _, result := range stored {
		require.Equal(t, INSTALLATION_ID, result.InstallationID)
		require.NotZero(t, result.CreatedAt)
		require.Contains(t, topics, result.Topic)
	}
}

func Test_Unsubscribe(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)
	svc := createService(t, db)

	err := svc.Subscribe(ctx, INSTALLATION_ID, []string{TOPIC})
	require.NoError(t, err)

	err = svc.Unsubscribe(ctx, INSTALLATION_ID, []string{TOPIC})
	require.NoError(t, err)

	stored := fetchSubscriptions(t, ctx, db, INSTALLATION_ID)
	require.Len(t, stored, 1)
	require.False(t, stored[0].IsActive)
}

func Test_UnsubscribeResubscribe(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)
	svc := createService(t, db)

	require.NoError(t, svc.Subscribe(ctx, INSTALLATION_ID, []string{TOPIC}))
	require.NoError(t, svc.Unsubscribe(ctx, INSTALLATION_ID, []string{TOPIC}))
	require.NoError(t, svc.Subscribe(ctx, INSTALLATION_ID, []string{TOPIC}))

	stored := fetchSubscriptions(t, ctx, db, INSTALLATION_ID)
	require.Len(t, stored, 1)
	require.True(t, stored[0].IsActive)
}

func Test_SubscribeWithMetadata(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)
	svc := createService(t, db)

	key := []byte("key")
	err := svc.SubscribeWithMetadata(ctx, INSTALLATION_ID, []interfaces.SubscriptionInput{{
		Topic:    TOPIC,
		IsSilent: true,
		HmacKeys: []interfaces.HmacKey{{
			ThirtyDayPeriodsSinceEpoch: 1,
			Key:                        key,
		}},
	}})
	require.NoError(t, err)

	results, err := svc.GetSubscriptions(ctx, TOPIC, 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, TOPIC, results[0].Topic)
	require.NotNil(t, results[0].HmacKey)
	require.True(t, results[0].IsSilent)
	require.Equal(t, key, results[0].HmacKey.Key)
}

func Test_UpdateIsSilent(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)
	svc := createService(t, db)

	require.NoError(t, svc.SubscribeWithMetadata(ctx, INSTALLATION_ID, []interfaces.SubscriptionInput{{
		Topic:    TOPIC,
		IsSilent: false,
	}}))

	results, err := svc.GetSubscriptions(ctx, TOPIC, 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.False(t, results[0].IsSilent)

	require.NoError(t, svc.SubscribeWithMetadata(ctx, INSTALLATION_ID, []interfaces.SubscriptionInput{{
		Topic:    TOPIC,
		IsSilent: true,
	}}))

	results, err = svc.GetSubscriptions(ctx, TOPIC, 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.True(t, results[0].IsSilent)
}

func Test_UpdateHmacKeys(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)
	svc := createService(t, db)

	require.NoError(t, svc.SubscribeWithMetadata(ctx, INSTALLATION_ID, []interfaces.SubscriptionInput{{
		Topic:    TOPIC,
		IsSilent: true,
		HmacKeys: []interfaces.HmacKey{{
			ThirtyDayPeriodsSinceEpoch: 1,
			Key:                        []byte("key"),
		}},
	}}))

	require.NoError(t, svc.SubscribeWithMetadata(ctx, INSTALLATION_ID, []interfaces.SubscriptionInput{{
		Topic:    TOPIC,
		IsSilent: true,
		HmacKeys: []interfaces.HmacKey{{
			ThirtyDayPeriodsSinceEpoch: 1,
			Key:                        []byte("key2"),
		}},
	}}))

	results, err := svc.GetSubscriptions(ctx, TOPIC, 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, []byte("key2"), results[0].HmacKey.Key)
}

func Test_GetSubscriptions(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)
	svc := createService(t, db)

	require.NoError(t, svc.Subscribe(ctx, INSTALLATION_ID, []string{TOPIC}))

	subs, err := svc.GetSubscriptions(ctx, TOPIC, 1)
	require.NoError(t, err)
	require.Len(t, subs, 1)
}
