package installations

import (
	"context"
	"database/sql"
	"time"

	"github.com/xmtp/example-notification-server-go/pkg/db/queries"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"go.uber.org/zap"
)

type DefaultInstallationService struct {
	logger  *zap.Logger
	db      *sql.DB
	queries *queries.Queries
}

func NewInstallationsService(logger *zap.Logger, db *sql.DB) *DefaultInstallationService {
	return &DefaultInstallationService{
		logger:  logger.Named("installations"),
		db:      db,
		queries: queries.New(db),
	}
}

func (s DefaultInstallationService) Register(
	ctx context.Context,
	installation interfaces.Installation,
) (res *interfaces.RegisterResponse, err error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	qtx := queries.New(tx)
	if err = qtx.UpsertInstallation(ctx, installation.Id); err != nil {
		return nil, err
	}

	updatedAt := installation.DeliveryMechanism.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	err = qtx.UpsertDeviceDeliveryMechanism(ctx, queries.UpsertDeviceDeliveryMechanismParams{
		InstallationID: installation.Id,
		Kind:           string(installation.DeliveryMechanism.Kind),
		Token:          installation.DeliveryMechanism.Token,
		UpdatedAt:      updatedAt,
	})
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &interfaces.RegisterResponse{
		InstallationId: installation.Id,
		ValidUntil:     getExpiry(updatedAt),
	}, nil
}

func (s DefaultInstallationService) Delete(ctx context.Context, installationID string) error {
	return s.queries.SoftDeleteInstallation(ctx, queries.SoftDeleteInstallationParams{
		ID:        installationID,
		DeletedAt: sql.NullTime{Time: time.Now(), Valid: true},
	})
}

func (s DefaultInstallationService) GetInstallations(
	ctx context.Context,
	installationIDs []string,
) ([]interfaces.Installation, error) {
	if len(installationIDs) == 0 {
		return []interfaces.Installation{}, nil
	}

	results, err := s.queries.GetLatestInstallations(ctx, installationIDs)
	if err != nil {
		return nil, err
	}

	out := make([]interfaces.Installation, 0, len(results))
	for _, result := range results {
		out = append(out, interfaces.Installation{
			Id: result.InstallationID,
			DeliveryMechanism: interfaces.DeliveryMechanism{
				Kind:      interfaces.DeliveryMechanismKind(result.Kind),
				Token:     result.Token,
				UpdatedAt: result.UpdatedAt,
			},
		})
	}
	return out, nil
}

func getExpiry(createdAt time.Time) time.Time {
	return createdAt
}
