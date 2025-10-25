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
	s.logger.Info("💾 Saving subscriptions to database",
		zap.String("installation_id", installationId),
		zap.Int("topic_count", len(topics)),
	)

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
			s.logger.Error("Error updating existing subscriptions", zap.Error(err))
			return err
		}

		s.logger.Info("Updated existing subscriptions",
			zap.Int("updated_count", len(out)),
		)

		topicMap := make(map[string]bool)
		for _, topic := range topics {
			topicMap[topic] = true
		}

		// Remove already updated results from the map
		for _, result := range out {
			delete(topicMap, result.Topic)
		}

		newSubsCount := len(topicMap)
		s.logger.Info("Creating new subscriptions",
			zap.Int("new_subscription_count", newSubsCount),
		)

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
				s.logger.Error("Error inserting new subscription",
					zap.String("topic", topic),
					zap.Error(err))
				return err
			}

			s.logger.Debug("Created subscription",
				zap.String("topic", topic),
				zap.String("installation_id", installationId),
			)
		}

		s.logger.Info("✅ All subscriptions saved successfully",
			zap.Int("updated", len(out)),
			zap.Int("created", newSubsCount),
			zap.Int("total", len(topics)),
		)

		return nil
	})
}

func (s SubscriptionsService) SubscribeWithMetadata(ctx context.Context, installationId string, subscriptions []interfaces.SubscriptionInput) error {
	s.logger.Info("💾 Saving subscriptions with metadata to database",
		zap.String("installation_id", installationId),
		zap.Int("subscription_count", len(subscriptions)),
	)

	return s.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		toUpdate := make([]*db.Subscription, len(subscriptions))
		for idx, sub := range subscriptions {
			s.logger.Debug("Preparing subscription for upsert",
				zap.String("topic", sub.Topic),
				zap.Bool("is_silent", sub.IsSilent),
				zap.Int("hmac_key_count", len(sub.HmacKeys)),
			)

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
			s.logger.Error("Error upserting subscriptions", zap.Error(err))
			return err
		}

		s.logger.Info("✅ Subscriptions upserted",
			zap.Int("count", len(updated)),
		)

		for _, u := range updated {
			s.logger.Debug("Upserted subscription",
				zap.Int64("id", u.Id),
				zap.String("topic", u.Topic),
			)
		}

		topicIdMap := makeTopicIdMap(updated)
		hmacKeyUpdates := []db.SubscriptionHmacKeys{}
		for _, sub := range subscriptions {
			subscriptionId, exists := topicIdMap[sub.Topic]
			if !exists {
				s.logger.Warn("⚠️ Skipping topic because subscription not found",
					zap.String("topic", sub.Topic))
				continue
			}

			s.logger.Debug("Processing HMAC keys for subscription",
				zap.String("topic", sub.Topic),
				zap.Int64("subscription_id", subscriptionId),
				zap.Int("hmac_key_count", len(sub.HmacKeys)),
			)

			for _, keyUpdate := range sub.HmacKeys {
				hmacKeyUpdates = append(hmacKeyUpdates, db.SubscriptionHmacKeys{
					SubscriptionId:             subscriptionId,
					ThirtyDayPeriodsSinceEpoch: int32(keyUpdate.ThirtyDayPeriodsSinceEpoch),
					Key:                        keyUpdate.Key,
				})
			}
		}

		if len(hmacKeyUpdates) > 0 {
			s.logger.Info("💾 Saving HMAC keys",
				zap.Int("hmac_key_count", len(hmacKeyUpdates)),
			)

			_, err = tx.NewInsert().
				Model(&hmacKeyUpdates).
				On("CONFLICT (subscription_id, thirty_day_periods_since_epoch) DO UPDATE").
				Set("key = EXCLUDED.key").
				Exec(ctx)

			if err != nil {
				s.logger.Error("Error saving HMAC keys", zap.Error(err))
				return err
			}

			s.logger.Info("✅ HMAC keys saved successfully")
		} else {
			s.logger.Info("No HMAC keys to save")
		}

		s.logger.Info("✅ SubscribeWithMetadata completed successfully",
			zap.String("installation_id", installationId),
			zap.Int("subscriptions", len(subscriptions)),
			zap.Int("hmac_keys", len(hmacKeyUpdates)),
		)

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
	s.logger.Debug("🔍 Querying subscriptions from database",
		zap.String("topic", topic),
		zap.Int("thirty_day_period", thirtyDayPeriod),
	)

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
		s.logger.Error("Error querying subscriptions", zap.Error(err))
		return nil, err
	}

	s.logger.Debug("Found subscriptions in database",
		zap.String("topic", topic),
		zap.Int("count", len(results)),
	)

	for _, result := range results {
		s.logger.Debug("Subscription found",
			zap.String("installation_id", result.InstallationId),
			zap.String("topic", result.Topic),
			zap.Bool("is_silent", result.IsSilent),
			zap.Bool("is_active", result.IsActive),
			zap.Int("hmac_key_count", len(result.HmacKeys)),
		)
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
