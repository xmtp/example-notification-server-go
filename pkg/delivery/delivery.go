package delivery

import (
	"context"

	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"go.uber.org/zap"
)

type DefaultDeliveryService struct {
	logger            *zap.Logger
	notificationTopic string
	apns              *ApnsDeliveryAgent
	fcm               *FcmDeliveryAgent
}

func NewDeliveryService(logger *zap.Logger, apns *ApnsDeliveryAgent, fcm *FcmDeliveryAgent) *DefaultDeliveryService {
	return &DefaultDeliveryService{
		logger: logger.Named("delivery-service"),
		apns:   apns,
		fcm:    fcm,
	}
}

func (d DefaultDeliveryService) Send(ctx context.Context, req interfaces.SendRequest) error {
	d.logger.Info("Sending notification", zap.Any("req", req))
	var err error
	for _, installation := range req.Installations {
		if installation.DeliveryMechanism.Kind == interfaces.APNS && d.apns != nil {
			// TODO: Better error handling for cases where one message succeeds and another fails
			err = d.apns.Send(ctx, installation.DeliveryMechanism.Token, req.Message.GetContentTopic(), string(req.Message.Message))
		}
		if installation.DeliveryMechanism.Kind == interfaces.FCM && d.fcm != nil {
			err = d.fcm.Send(ctx, installation.DeliveryMechanism.Token, req.Message.GetContentTopic(), string(req.Message.Message))
		}
		if err != nil {
			break
		}
	}

	return err
}
