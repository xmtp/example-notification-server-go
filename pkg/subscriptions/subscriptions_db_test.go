package subscriptions

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xmtp/example-notification-server-go/pkg/db"
	"github.com/xmtp/example-notification-server-go/pkg/topics"
	"github.com/xmtp/example-notification-server-go/test"
)

func Test_PopulateTopicIDs(t *testing.T) {

	var (
		ctx    = t.Context()
		testDB = test.CreateTestDb(t)
	)

	t.Run("topic ID added for old group messages", func(t *testing.T) {

		var (
			installationID = "installation-id"
			topicID        = "generic-group-topic-id"
			topic          = topics.V3GroupPrefix + topicID

			sub = db.Subscription{
				InstallationId: installationID,
				Topic:          topic,
			}
		)

		_, err := testDB.NewInsert().Model(&sub).Exec(ctx)
		require.NoError(t, err)

		// Confirm topic ID is initally empty.
		var read db.Subscription
		err = testDB.NewSelect().Where("topic = ?", topic).Model(&read).Scan(ctx)
		require.NoError(t, err)

		require.Equal(t, installationID, read.InstallationId)
		require.Equal(t, topic, read.Topic)

		require.Empty(t, read.TopicID)

		// Invoke the function to populate topicIDs.
		_, err = testDB.QueryContext(ctx, "SELECT populate_topic_ids();")
		require.NoError(t, err)

		// Verify the row was updated and has a valid topic ID.
		err = testDB.NewSelect().Where("topic = ?", topic).Model(&read).Scan(ctx)
		require.NoError(t, err)

		require.Equal(t, installationID, read.InstallationId)
		require.Equal(t, topic, read.Topic)
		require.Equal(t, topicID, read.TopicID)
	})
	t.Run("topic ID added for old welcome messages", func(t *testing.T) {

		var (
			installationID = "installation-id"
			topicID        = "generic-welcome-topic-id"
			topic          = topics.V3WelcomeMessagePrefix + topicID

			sub = db.Subscription{
				InstallationId: installationID,
				Topic:          topic,
			}
		)

		_, err := testDB.NewInsert().Model(&sub).Exec(ctx)
		require.NoError(t, err)

		// Confirm topic ID is initally empty.
		var read db.Subscription
		err = testDB.NewSelect().Where("topic = ?", topic).Model(&read).Scan(ctx)
		require.NoError(t, err)

		require.Equal(t, installationID, read.InstallationId)
		require.Equal(t, topic, read.Topic)

		require.Empty(t, read.TopicID)

		// Invoke the function to populate topicIDs.
		_, err = testDB.QueryContext(ctx, "SELECT populate_topic_ids();")
		require.NoError(t, err)

		// Verify the row was updated and has a valid topic ID.
		err = testDB.NewSelect().Where("topic = ?", topic).Model(&read).Scan(ctx)
		require.NoError(t, err)

		require.Equal(t, installationID, read.InstallationId)
		require.Equal(t, topic, read.Topic)
		require.Equal(t, topicID, read.TopicID)
	})
}

func Test_DisableInvalidTopics(t *testing.T) {

	var (
		ctx    = t.Context()
		testDB = test.CreateTestDb(t)

		installationID = "installationID"

		validTopics = []string{
			"/xmtp/mls/1/g-topic1",
			"/xmtp/mls/1/g-topic2",
			"/xmtp/mls/1/g-topic3",
			"/xmtp/mls/1/w-topic1",
			"/xmtp/mls/1/w-topic2",
		}
		invalidTopics = []string{
			"/xmtp/mls/1/f-topic1",
			"/xmtp/mls/1/s-topic1",
			"/xmtp/mls/0/something-random",
			"whatever",
		}

		topics = slices.Concat(validTopics, invalidTopics)
	)

	subs := make([]db.Subscription, len(topics))
	for i := range topics {
		subs[i] = db.Subscription{
			InstallationId: installationID,
			Topic:          topics[i],
			TopicID:        topics[i] + "dummy", // Kludge: Not a realistic scenario but make the insert work so we test disabling.
		}
	}

	res, err := testDB.NewInsert().Model(&subs).Exec(ctx)
	require.NoError(t, err)

	rows, err := res.RowsAffected()
	require.NoError(t, err)
	require.EqualValues(t, len(subs), rows)
}

func Test_IdenticalSubscriptionsDisallowed(t *testing.T) {

	var (
		ctx    = t.Context()
		testDB = test.CreateTestDb(t)
	)

	t.Run("duplicate topic ID fails", func(t *testing.T) {

		var (
			installationID = "installation-id"
			topicID        = "generic-group-topic-id"

			sub = db.Subscription{
				InstallationId: installationID,
				TopicID:        topicID,
			}
		)

		_, err := testDB.NewInsert().Model(&sub).Exec(ctx)
		require.NoError(t, err)

		dup := db.Subscription{
			InstallationId: sub.InstallationId,
			TopicID:        sub.TopicID,
		}

		_, err = testDB.NewInsert().Model(&dup).Exec(ctx)
		require.ErrorContains(t, err, "subscriptions_installation_id_topic_id_idx") // duplicate instID + topicID combination
	})
}
