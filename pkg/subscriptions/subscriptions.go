package subscriptions

import (
	"context"
	"database/sql"

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

func (s SubscriptionsService) Subscribe(ctx context.Context, installationID string, topics []string) (err error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	qtx := queries.New(tx)
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

	for topic := range topicMap {
		err = qtx.InsertSubscription(ctx, queries.InsertSubscriptionParams{
			InstallationID: installationID,
			Topic:          topic,
			IsActive:       true,
			IsSilent:       false,
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s SubscriptionsService) SubscribeWithMetadata(
	ctx context.Context,
	installationID string,
	subscriptions []interfaces.SubscriptionInput,
) (err error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	qtx := queries.New(tx)
	for _, sub := range subscriptions {
		row, err := qtx.UpsertSubscription(ctx, queries.UpsertSubscriptionParams{
			InstallationID: installationID,
			Topic:          sub.Topic,
			IsSilent:       sub.IsSilent,
		})
		if err != nil {
			return err
		}

		for _, keyUpdate := range sub.HmacKeys {
			err = qtx.UpsertSubscriptionHmacKey(ctx, queries.UpsertSubscriptionHmacKeyParams{
				SubscriptionID:             row.ID,
				ThirtyDayPeriodsSinceEpoch: int32(keyUpdate.ThirtyDayPeriodsSinceEpoch),
				Key:                        keyUpdate.Key,
			})
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
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
