package subscriptions

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/xmtp/example-notification-server-go/pkg/db"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"go.uber.org/zap"
)

type SubscriptionsService struct {
	logger *zap.Logger
	db     *bun.DB
}

func NewSubscriptionsService(logger *zap.Logger, db *bun.DB) *SubscriptionsService {
	return &SubscriptionsService{
		logger: logger.Named("subscriptions-service"),
		db:     db,
	}
}

func (s SubscriptionsService) Subscribe(ctx context.Context, installationId string, topics []string) error {
	return s.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		out := make([]db.Subscription, 0)
		// Update any existing results
		_, err := tx.NewUpdate().
			Model(&out).
			Where("installation_id = ?", installationId).
			Where("topic IN (?)", bun.In(topics)).
			Set("is_active = ?", true).
			Returning("topic").
			Exec(ctx)

		if err != nil {
			return err
		}

		topicMap := make(map[string]bool)
		for _, topic := range topics {
			topicMap[topic] = true
		}

		// Remove already updated results from the map
		for _, result := range out {
			delete(topicMap, result.Topic)
		}

		for topic := range topicMap {
			newSub := db.Subscription{
				InstallationId: installationId,
				Topic:          topic,
				IsActive:       true,
			}
			_, err = tx.NewInsert().
				Model(&newSub).
				Exec(ctx)

			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (s SubscriptionsService) SubscribeWithMetadata(ctx context.Context, installationId string, subscriptions []interfaces.SubscriptionInput) error {
	return s.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		toUpdate := make([]*db.Subscription, len(subscriptions))
		for idx, sub := range subscriptions {
			toUpdate[idx] = &db.Subscription{
				InstallationId: installationId,
				Topic:          sub.Topic,
				IsActive:       true,
				IsSilent:       sub.IsSilent,
			}
		}

		updated := make([]*db.Subscription, 0)
		_, err := tx.NewInsert().
			Model(&toUpdate).
			On("CONFLICT (installation_id, topic) DO UPDATE").
			Set("is_active = true").
			Set("is_silent = EXCLUDED.is_silent").
			Returning("id, topic").
			Exec(ctx, &updated)

		if err != nil {
			return err
		}

		topicIdMap := makeTopicIdMap(updated)
		hmacKeyUpdates := []db.SubscriptionHmacKeys{}
		for _, sub := range subscriptions {
			subscriptionId, exists := topicIdMap[sub.Topic]
			if !exists {
				s.logger.Info("Skipping topic because subscription not found", zap.String("topic", sub.Topic))
				continue
			}
			for _, keyUpdate := range sub.HmacKeys {
				hmacKeyUpdates = append(hmacKeyUpdates, db.SubscriptionHmacKeys{
					SubscriptionId:             subscriptionId,
					ThirtyDayPeriodsSinceEpoch: int32(keyUpdate.ThirtyDayPeriodsSinceEpoch),
					Key:                        keyUpdate.Key,
				})
			}
		}

		if len(hmacKeyUpdates) > 0 {
			_, err = tx.NewInsert().
				Model(&hmacKeyUpdates).
				On("CONFLICT (subscription_id, thirty_day_periods_since_epoch) DO UPDATE").
				Set("key = EXCLUDED.key").
				Exec(ctx)
		}

		return err
	})
}

func (s SubscriptionsService) Unsubscribe(ctx context.Context, installationId string, topics []string) error {
	return s.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewUpdate().
			Model((*db.Subscription)(nil)).
			Where("installation_id = ?", installationId).
			Where("topic IN (?)", bun.In(topics)).
			Set("is_active = ?", false).
			Exec(ctx)

		return err
	})
}

func (s SubscriptionsService) GetSubscriptions(ctx context.Context, topic string, thirtyDayPeriod int) (out []interfaces.Subscription, err error) {
	results := make([]db.Subscription, 0)
	err = s.db.NewSelect().
		Model(&results).
		Where("topic = ?", topic).
		Where("is_active = TRUE").
		Relation("HmacKeys", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("thirty_day_periods_since_epoch = ?", thirtyDayPeriod)
		}).
		Scan(ctx)

	if err != nil {
		return nil, err
	}
	// s.logger.Info("Results", zap.Any("results", results))
	for _, result := range results {
		out = append(out, transformResult(result))
	}

	return out, err
}

func makeTopicIdMap(subscriptions []*db.Subscription) map[string]int64 {
	out := make(map[string]int64)
	for _, sub := range subscriptions {
		out[sub.Topic] = sub.Id
	}
	return out
}

func transformResult(dbSubscription db.Subscription) interfaces.Subscription {
	return interfaces.Subscription{
		Id:             dbSubscription.Id,
		CreatedAt:      dbSubscription.CreatedAt,
		InstallationId: dbSubscription.InstallationId,
		Topic:          dbSubscription.Topic,
		IsActive:       dbSubscription.IsActive,
		IsSilent:       dbSubscription.IsSilent,
		HmacKey:        extractHmacKey(dbSubscription.HmacKeys),
	}
}

func extractHmacKey(dbKeys []*db.SubscriptionHmacKeys) *interfaces.HmacKey {
	for _, key := range dbKeys {
		return &interfaces.HmacKey{
			ThirtyDayPeriodsSinceEpoch: int(key.ThirtyDayPeriodsSinceEpoch),
			Key:                        key.Key,
		}
	}
	return nil
}
