package subscriptions

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/xmtp/example-notification-server-go/pkg/db"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"go.uber.org/zap"
)

type Service struct {
	logger *zap.Logger
	db     *bun.DB
}

func NewService(logger *zap.Logger, db *bun.DB) *Service {
	return &Service{
		logger: logger.Named("subscriptions-service"),
		db:     db,
	}
}

func (s Service) Subscribe(ctx context.Context, installationId string, topicIDs []string) error {
	return s.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {

		// TODO: Do a single upsert.

		out := make([]db.Subscription, 0)
		// Update any existing results
		_, err := tx.NewUpdate().
			Model(&out).
			Where("installation_id = ?", installationId).
			Where("topic_id IN (?)", bun.In(topicIDs)).
			Set("is_active = ?", true).
			Returning("topic_id").
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("could not update subscriptions: %w", err)
		}

		topicIDMap := make(map[string]bool)
		for _, topicID := range topicIDs {
			topicIDMap[topicID] = true
		}

		// Remove already updated results from the map
		for _, result := range out {
			delete(topicIDMap, result.TopicID)
		}

		if len(topicIDMap) == 0 {
			return nil
		}

		newSubs := make([]db.Subscription, 0, len(topicIDs))
		for topicID := range topicIDMap {
			newSub := db.Subscription{
				InstallationId: installationId,
				Topic:          topicID, // Keep this to satisfy constraints, though in the future we might remove this.
				TopicID:        topicID,
				IsActive:       true,
			}

			newSubs = append(newSubs, newSub)
		}

		// Bulk insert new subs.
		_, err = tx.NewInsert().Model(&newSubs).Exec(ctx)
		if err != nil {
			return fmt.Errorf("could not insert subscription: %w", err)
		}

		return nil
	})
}

func (s Service) SubscribeWithMetadata(ctx context.Context, installationId string, subscriptions []interfaces.SubscriptionInput) error {
	return s.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		toUpdate := make([]*db.Subscription, len(subscriptions))
		for i, sub := range subscriptions {
			toUpdate[i] = &db.Subscription{
				InstallationId: installationId,
				Topic:          sub.TopicID,
				TopicID:        sub.TopicID,
				IsActive:       true,
				IsSilent:       sub.IsSilent,
			}
		}

		updated := make([]*db.Subscription, 0)
		_, err := tx.NewInsert().
			Model(&toUpdate).
			On("CONFLICT (installation_id, topic_id) DO UPDATE").
			Set("is_active = true").
			Set("is_silent = EXCLUDED.is_silent").
			Returning("id, topic_id").
			Exec(ctx, &updated)

		if err != nil {
			return err
		}

		topicIdMap := makeTopicIdMap(updated)
		hmacKeyUpdates := []db.SubscriptionHmacKeys{}

		for _, sub := range subscriptions {
			subscriptionId, exists := topicIdMap[sub.TopicID]
			if !exists {
				s.logger.Info("Skipping topic because subscription not found", zap.String("topic", sub.TopicID))
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

		if len(hmacKeyUpdates) == 0 {
			return nil
		}

		_, err = tx.NewInsert().
			Model(&hmacKeyUpdates).
			On("CONFLICT (subscription_id, thirty_day_periods_since_epoch) DO UPDATE").
			Set("key = EXCLUDED.key").
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("could not update hmac key data: %w", err)
		}

		return err
	})
}

func (s Service) Unsubscribe(ctx context.Context, installationId string, topicIDs []string) error {
	return s.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewUpdate().
			Model((*db.Subscription)(nil)).
			Where("installation_id = ?", installationId).
			Where("topic_id IN (?)", bun.In(topicIDs)).
			Set("is_active = ?", false).
			Exec(ctx)

		return err
	})
}

func (s Service) GetSubscriptions(ctx context.Context, topicID string, thirtyDayPeriod int) ([]interfaces.Subscription, error) {

	results := make([]db.Subscription, 0)
	err := s.db.NewSelect().
		Model(&results).
		Where("topic_id = ?", topicID).
		Where("is_active = TRUE").
		Relation("HmacKeys", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("thirty_day_periods_since_epoch = ?", thirtyDayPeriod)
		}).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	out := make([]interfaces.Subscription, len(results))
	// s.logger.Info("Results", zap.Any("results", results))
	for i := range results {
		out[i] = transformResult(results[i])
	}

	return out, err
}

func makeTopicIdMap(subscriptions []*db.Subscription) map[string]int64 {
	out := make(map[string]int64)
	for _, sub := range subscriptions {
		out[sub.TopicID] = sub.Id
	}
	return out
}

func transformResult(s db.Subscription) interfaces.Subscription {
	return interfaces.Subscription{
		Id:             s.Id,
		CreatedAt:      s.CreatedAt,
		InstallationId: s.InstallationId,
		Topic:          s.Topic,
		TopicID:        s.TopicID,
		IsActive:       s.IsActive,
		IsSilent:       s.IsSilent,
		HmacKey:        extractHmacKey(s.HmacKeys),
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
