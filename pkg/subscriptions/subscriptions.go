package subscriptions

import (
	"context"
	"database/sql"

	database "github.com/xmtp/example-notification-server-go/pkg/db"
	"github.com/xmtp/example-notification-server-go/pkg/db/queries"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"go.uber.org/zap"
)

type SubscriptionsService struct {
	logger  *zap.Logger
	db      *sql.DB
	queries *queries.Queries
}

func NewSubscriptionsService(logger *zap.Logger, db *sql.DB) *SubscriptionsService {
	return &SubscriptionsService{
		logger:  logger.Named("subscriptions-service"),
		db:      db,
		queries: queries.New(db),
	}
}

func (s SubscriptionsService) Subscribe(ctx context.Context, installationID string, topics []string) error {
	return database.RunInTx(ctx, s.db, func(qtx *queries.Queries) error {
		updated, err := qtx.ReactivateSubscriptions(ctx, queries.ReactivateSubscriptionsParams{
			InstallationID: installationID,
			Topics:         topics,
		})
		if err != nil {
			return err
		}

		topicMap := make(map[string]bool, len(topics))
		for _, topic := range topics {
			topicMap[topic] = true
		}
		for _, result := range updated {
			delete(topicMap, result.Topic)
		}

		remaining := make([]string, 0, len(topicMap))
		for topic := range topicMap {
			remaining = append(remaining, topic)
		}

		if len(remaining) > 0 {
			return qtx.BatchInsertSubscriptions(ctx, queries.BatchInsertSubscriptionsParams{
				InstallationID: installationID,
				Topics:         remaining,
			})
		}

		return nil
	})
}

func (s SubscriptionsService) SubscribeWithMetadata(
	ctx context.Context,
	installationID string,
	subscriptions []interfaces.SubscriptionInput,
) error {
	return database.RunInTx(ctx, s.db, func(qtx *queries.Queries) error {
		if len(subscriptions) == 0 {
			return nil
		}

		topics := make([]string, len(subscriptions))
		isSilents := make([]bool, len(subscriptions))
		for i, sub := range subscriptions {
			topics[i] = sub.Topic
			isSilents[i] = sub.IsSilent
		}

		rows, err := qtx.BatchUpsertSubscriptions(ctx, queries.BatchUpsertSubscriptionsParams{
			InstallationID: installationID,
			Topics:         topics,
			IsSilents:      isSilents,
		})
		if err != nil {
			return err
		}

		topicToID := make(map[string]int64, len(rows))
		for _, row := range rows {
			topicToID[row.Topic] = row.ID
		}

		var (
			subscriptionIDs []int64
			periods         []int32
			keys            [][]byte
		)
		for _, sub := range subscriptions {
			id := topicToID[sub.Topic]
			for _, keyUpdate := range sub.HmacKeys {
				subscriptionIDs = append(subscriptionIDs, id)
				periods = append(periods, int32(keyUpdate.ThirtyDayPeriodsSinceEpoch))
				keys = append(keys, keyUpdate.Key)
			}
		}

		if len(subscriptionIDs) > 0 {
			return qtx.BatchUpsertSubscriptionHmacKeys(ctx, queries.BatchUpsertSubscriptionHmacKeysParams{
				SubscriptionIds: subscriptionIDs,
				Periods:         periods,
				Keys:            keys,
			})
		}

		return nil
	})
}

func (s SubscriptionsService) Unsubscribe(ctx context.Context, installationID string, topics []string) error {
	return s.queries.DeactivateSubscriptions(ctx, queries.DeactivateSubscriptionsParams{
		InstallationID: installationID,
		Topics:         topics,
	})
}

func (s SubscriptionsService) GetSubscriptions(
	ctx context.Context,
	topic string,
	thirtyDayPeriod int,
) ([]interfaces.Subscription, error) {
	results, err := s.queries.ListActiveSubscriptionsByTopicAndPeriod(
		ctx,
		queries.ListActiveSubscriptionsByTopicAndPeriodParams{
			ThirtyDayPeriod: int32(thirtyDayPeriod),
			Topic:           topic,
		},
	)
	if err != nil {
		return nil, err
	}

	out := make([]interfaces.Subscription, 0, len(results))
	for _, result := range results {
		subscription := interfaces.Subscription{
			Id:             result.ID,
			CreatedAt:      result.CreatedAt,
			InstallationId: result.InstallationID,
			Topic:          result.Topic,
			IsActive:       result.IsActive,
			IsSilent:       result.IsSilent,
		}
		if result.ThirtyDayPeriodsSinceEpoch.Valid {
			subscription.HmacKey = &interfaces.HmacKey{
				ThirtyDayPeriodsSinceEpoch: int(result.ThirtyDayPeriodsSinceEpoch.Int32),
				Key:                        result.Key,
			}
		}
		out = append(out, subscription)
	}

	return out, nil
}
