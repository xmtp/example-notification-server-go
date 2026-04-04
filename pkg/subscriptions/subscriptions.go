package subscriptions

import (
	"context"
	"database/sql"
	"errors"

	database "github.com/xmtp/example-notification-server-go/pkg/db"
	"github.com/xmtp/example-notification-server-go/pkg/db/queries"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	topicutil "github.com/xmtp/example-notification-server-go/pkg/topics"
	"github.com/xmtp/xmtpd/pkg/topic"
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

func (s SubscriptionsService) Subscribe(ctx context.Context, installationID string, topics []*topic.Topic) error {
	return database.RunInTx(ctx, s.db, func(qtx *queries.Queries) error {
		topicBytes := make([][]byte, len(topics))
		for i, t := range topics {
			topicBytes[i] = t.Bytes()
		}

		updated, err := qtx.ReactivateSubscriptions(ctx, queries.ReactivateSubscriptionsParams{
			InstallationID: installationID,
			Topics:         topicBytes,
		})
		if err != nil {
			return err
		}

		reactivated := make(map[string]bool, len(updated))
		for _, result := range updated {
			reactivated[string(result.Topic)] = true
		}

		var remaining [][]byte
		for _, t := range topics {
			if !reactivated[string(t.Bytes())] {
				remaining = append(remaining, t.Bytes())
			}
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

		topicBytes := make([][]byte, len(subscriptions))
		isSilents := make([]bool, len(subscriptions))
		for i, sub := range subscriptions {
			if sub.Topic == nil {
				return errors.New("subscription topic must not be nil")
			}
			topicBytes[i] = sub.Topic.Bytes()
			isSilents[i] = sub.IsSilent
		}

		rows, err := qtx.BatchUpsertSubscriptions(ctx, queries.BatchUpsertSubscriptionsParams{
			InstallationID: installationID,
			Topics:         topicBytes,
			IsSilents:      isSilents,
		})
		if err != nil {
			return err
		}

		topicToID := make(map[string]int64, len(rows))
		for _, row := range rows {
			topicToID[string(row.Topic)] = row.ID
		}

		var (
			subscriptionIDs []int64
			periods         []int32
			keys            [][]byte
		)
		for _, sub := range subscriptions {
			id := topicToID[string(sub.Topic.Bytes())]
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

func (s SubscriptionsService) Unsubscribe(ctx context.Context, installationID string, topics []*topic.Topic) error {
	topicBytes := make([][]byte, len(topics))
	for i, t := range topics {
		topicBytes[i] = t.Bytes()
	}
	return s.queries.DeactivateSubscriptions(ctx, queries.DeactivateSubscriptionsParams{
		InstallationID: installationID,
		Topics:         topicBytes,
	})
}

func (s SubscriptionsService) GetSubscriptions(
	ctx context.Context,
	t *topic.Topic,
	thirtyDayPeriod int,
) ([]interfaces.Subscription, error) {
	if t == nil {
		return nil, errors.New("topic must not be nil")
	}
	results, err := s.queries.ListActiveSubscriptionsByTopicAndPeriod(
		ctx,
		queries.ListActiveSubscriptionsByTopicAndPeriodParams{
			ThirtyDayPeriod: int32(thirtyDayPeriod),
			Topic:           t.Bytes(),
		},
	)
	if err != nil {
		return nil, err
	}

	out := make([]interfaces.Subscription, 0, len(results))
	for _, result := range results {
		parsedTopic, err := topic.ParseTopic(result.Topic)
		if err != nil {
			s.logger.Warn("failed to parse topic from DB", zap.Error(err))
			continue
		}
		subscription := interfaces.Subscription{
			Id:             result.ID,
			CreatedAt:      result.CreatedAt,
			InstallationId: result.InstallationID,
			Topic:          topicutil.TopicToString(parsedTopic),
			TopicV4:        parsedTopic,
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
