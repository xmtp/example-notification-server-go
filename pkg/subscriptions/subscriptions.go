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

func (s SubscriptionsService) GetSubscriptions(ctx context.Context, topic string) (out []interfaces.Subscription, err error) {
	results := make([]db.Subscription, 0)
	_, err = s.db.NewSelect().
		Model(&results).
		Where("topic = ?", topic).
		Where("is_active = ?", true).
		Exec(ctx)

	if err != nil {
		return nil, err
	}

	for _, result := range results {
		out = append(out, transformResult(result))
	}

	return out, err
}

func transformResult(dbSubscription db.Subscription) interfaces.Subscription {
	return interfaces.Subscription{
		Id:             dbSubscription.Id,
		CreatedAt:      dbSubscription.CreatedAt,
		InstallationId: dbSubscription.InstallationId,
		Topic:          dbSubscription.Topic,
		IsActive:       dbSubscription.IsActive,
	}
}
