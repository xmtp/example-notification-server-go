package subscriptions

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/xmtp/example-notification-server-go/pkg/db"
	"go.uber.org/zap"
)

type SubscriptionsService struct {
	logger *zap.Logger
	db     *bun.DB
}

func NewSubscriptionsService(logger *zap.Logger, db *bun.DB) *SubscriptionsService {
	return &SubscriptionsService{
		logger: logger,
		db:     db,
	}
}

func (s SubscriptionsService) Subscribe(ctx context.Context, installationId string, topics []string) error {
	return s.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		// Update any existing results
		res, err := s.db.NewUpdate().
			Model((*db.Subscription)(nil)).
			Where("installation_id = ?", installationId).
			Where("topic IN (?)", bun.In(topics)).
			Exec(ctx)
	})
}
