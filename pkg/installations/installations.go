package installations

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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
		logger: logger.Named("installations"),
		db:     db,
	}
}

func (s Service) Register(ctx context.Context, installation interfaces.Installation) (res *interfaces.RegisterResponse, err error) {
	err = s.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		// Not sure how I want to handle register calls to deleted installation_ids
		_, err := tx.NewInsert().
			Model(&db.Installation{
				Id: installation.Id,
			}).
			On("CONFLICT(id) DO UPDATE").
			Set("deleted_at = NULL").
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("could not save installation: %w", err)
		}

		_, err = tx.NewInsert().
			Model(&db.DeviceDeliveryMechanism{
				Kind:           installation.DeliveryMechanism.Kind,
				InstallationId: installation.Id,
				Token:          installation.DeliveryMechanism.Token,
				UpdatedAt:      installation.DeliveryMechanism.UpdatedAt,
			}).
			On("CONFLICT(installation_id, kind, token) DO UPDATE").
			// Set updated_at to provided value if already exists
			Set("updated_at = EXCLUDED.updated_at").Exec(ctx)

		if err != nil {
			return fmt.Errorf("could not save delivery mechanism: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &interfaces.RegisterResponse{
		InstallationId: installation.Id,
		ValidUntil:     getExpiry(installation.DeliveryMechanism.UpdatedAt),
	}, nil
}

func (s Service) Delete(ctx context.Context, installationId string) error {
	deletedAt := time.Now()
	installation := &db.Installation{Id: installationId, DeletedAt: &deletedAt}
	_, err := s.db.NewUpdate().
		Model(installation).
		Column("deleted_at").
		WherePK().
		Exec(ctx)

	return err
}

func (s Service) GetInstallations(ctx context.Context, installationIds []string) ([]interfaces.Installation, error) {
	// Abort if empty
	if len(installationIds) == 0 {
		return []interfaces.Installation{}, nil
	}

	results := make([]db.DeviceDeliveryMechanism, 0)
	err := s.db.NewSelect().
		Model((*db.DeviceDeliveryMechanism)(nil)).
		Where("installation_id IN (?)", bun.In(installationIds)).
		Where("installation.deleted_at IS NULL").
		Relation("Installation").
		DistinctOn("installation_id").
		Order("installation_id DESC").
		Order("updated_at DESC").
		Scan(ctx, &results)

	if err != nil {
		return nil, err
	}

	out := make([]interfaces.Installation, len(results))
	for i := range results {
		t := transformResult(results[i])
		out[i] = *t
	}

	return out, nil
}

func transformResult(m db.DeviceDeliveryMechanism) *interfaces.Installation {
	return &interfaces.Installation{
		Id: m.InstallationId,
		DeliveryMechanism: interfaces.DeliveryMechanism{
			Kind:      m.Kind,
			Token:     m.Token,
			UpdatedAt: m.UpdatedAt,
		},
	}
}

func getExpiry(createdAt time.Time) time.Time {
	// TODO: Figure out expiry time
	return createdAt
}
