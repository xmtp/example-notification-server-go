package delivery

import (
	"context"

	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"go.uber.org/zap"
)

type DefaultDeliveryService struct {
	logger *zap.Logger
}

func NewDeliveryService(logger *zap.Logger) *DefaultDeliveryService {
	return &DefaultDeliveryService{
		logger: logger.Named("delivery-serive"),
	}
}

func (d DefaultDeliveryService) Send(ctx context.Context, req interfaces.SendRequest) error {
	d.logger.Info("Sending notification", zap.Any("req", req))
	return nil
}
