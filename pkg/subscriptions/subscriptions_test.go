package subscriptions

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/testutils"
	topicutil "github.com/xmtp/example-notification-server-go/pkg/topics"
	topicpkg "github.com/xmtp/xmtpd/pkg/topic"
)

const INSTALLATION_ID = "installation_1"

// Valid V3 topic and its parsed form for testing
var TEST_TOPIC, _ = topicutil.ParseV3Topic("/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto")

type storedSubscription struct {
	ID             int64
	CreatedAt      time.Time
	InstallationID string
	Topic          []byte
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

	err := svc.Subscribe(ctx, INSTALLATION_ID, []*topicpkg.Topic{TEST_TOPIC})
	require.NoError(t, err)

	stored := fetchSubscriptions(t, ctx, db, INSTALLATION_ID)
	require.Len(t, stored, 1)
	require.Equal(t, INSTALLATION_ID, stored[0].InstallationID)
	require.True(t, stored[0].IsActive)
	require.Equal(t, TEST_TOPIC.Bytes(), stored[0].Topic)
}

func Test_SubscribeMultiple(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)
	svc := createService(t, db)

	topic1 := topicpkg.NewTopic(topicpkg.TopicKindGroupMessagesV1, []byte{0x01})
	topic2 := topicpkg.NewTopic(topicpkg.TopicKindGroupMessagesV1, []byte{0x02})
	topic3 := topicpkg.NewTopic(topicpkg.TopicKindGroupMessagesV1, []byte{0x03})
	topics := []*topicpkg.Topic{topic1, topic2, topic3}

	err := svc.Subscribe(ctx, INSTALLATION_ID, topics)
	require.NoError(t, err)

	stored := fetchSubscriptions(t, ctx, db, INSTALLATION_ID)
	require.Len(t, stored, 3)

	storedTopicBytes := make([][]byte, len(stored))
	for i, result := range stored {
		require.Equal(t, INSTALLATION_ID, result.InstallationID)
		require.NotZero(t, result.CreatedAt)
		storedTopicBytes[i] = result.Topic
	}
	for _, tp := range topics {
		require.Contains(t, storedTopicBytes, tp.Bytes())
	}
}

func Test_Unsubscribe(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)
	svc := createService(t, db)

	err := svc.Subscribe(ctx, INSTALLATION_ID, []*topicpkg.Topic{TEST_TOPIC})
	require.NoError(t, err)

	err = svc.Unsubscribe(ctx, INSTALLATION_ID, []*topicpkg.Topic{TEST_TOPIC})
	require.NoError(t, err)

	stored := fetchSubscriptions(t, ctx, db, INSTALLATION_ID)
	require.Len(t, stored, 1)
	require.False(t, stored[0].IsActive)
}

func Test_UnsubscribeResubscribe(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)
	svc := createService(t, db)

	require.NoError(t, svc.Subscribe(ctx, INSTALLATION_ID, []*topicpkg.Topic{TEST_TOPIC}))
	require.NoError(t, svc.Unsubscribe(ctx, INSTALLATION_ID, []*topicpkg.Topic{TEST_TOPIC}))
	require.NoError(t, svc.Subscribe(ctx, INSTALLATION_ID, []*topicpkg.Topic{TEST_TOPIC}))

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
		Topic:    TEST_TOPIC,
		IsSilent: true,
		HmacKeys: []interfaces.HmacKey{{
			ThirtyDayPeriodsSinceEpoch: 1,
			Key:                        key,
		}},
	}})
	require.NoError(t, err)

	results, err := svc.GetSubscriptions(ctx, TEST_TOPIC, 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, TEST_TOPIC.Kind(), results[0].TopicV4.Kind())
	require.Equal(t, TEST_TOPIC.Bytes(), results[0].TopicV4.Bytes())
	require.NotNil(t, results[0].HmacKey)
	require.True(t, results[0].IsSilent)
	require.Equal(t, key, results[0].HmacKey.Key)
}

func Test_UpdateIsSilent(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)
	svc := createService(t, db)

	require.NoError(t, svc.SubscribeWithMetadata(ctx, INSTALLATION_ID, []interfaces.SubscriptionInput{{
		Topic:    TEST_TOPIC,
		IsSilent: false,
	}}))

	results, err := svc.GetSubscriptions(ctx, TEST_TOPIC, 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.False(t, results[0].IsSilent)

	require.NoError(t, svc.SubscribeWithMetadata(ctx, INSTALLATION_ID, []interfaces.SubscriptionInput{{
		Topic:    TEST_TOPIC,
		IsSilent: true,
	}}))

	results, err = svc.GetSubscriptions(ctx, TEST_TOPIC, 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.True(t, results[0].IsSilent)
}

func Test_UpdateHmacKeys(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)
	svc := createService(t, db)

	require.NoError(t, svc.SubscribeWithMetadata(ctx, INSTALLATION_ID, []interfaces.SubscriptionInput{{
		Topic:    TEST_TOPIC,
		IsSilent: true,
		HmacKeys: []interfaces.HmacKey{{
			ThirtyDayPeriodsSinceEpoch: 1,
			Key:                        []byte("key"),
		}},
	}}))

	require.NoError(t, svc.SubscribeWithMetadata(ctx, INSTALLATION_ID, []interfaces.SubscriptionInput{{
		Topic:    TEST_TOPIC,
		IsSilent: true,
		HmacKeys: []interfaces.HmacKey{{
			ThirtyDayPeriodsSinceEpoch: 1,
			Key:                        []byte("key2"),
		}},
	}}))

	results, err := svc.GetSubscriptions(ctx, TEST_TOPIC, 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, []byte("key2"), results[0].HmacKey.Key)
}

func Test_GetSubscriptions(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)
	svc := createService(t, db)

	require.NoError(t, svc.Subscribe(ctx, INSTALLATION_ID, []*topicpkg.Topic{TEST_TOPIC}))

	subs, err := svc.GetSubscriptions(ctx, TEST_TOPIC, 1)
	require.NoError(t, err)
	require.Len(t, subs, 1)
}

func Test_SubscribeWithMetadata_NilTopic(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)
	svc := createService(t, db)

	err := svc.SubscribeWithMetadata(ctx, INSTALLATION_ID, []interfaces.SubscriptionInput{{
		Topic:    nil,
		IsSilent: false,
	}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "nil")
}
