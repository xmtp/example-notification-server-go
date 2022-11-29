package delivery

import (
	"context"
	"encoding/base64"

	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"go.uber.org/zap"
)

type DefaultDeliveryService struct {
	logger            *zap.Logger
	notificationTopic string
	apns              *ApnsDelivery
	fcm               *FcmDelivery
}

func NewDeliveryService(logger *zap.Logger, apns *ApnsDelivery, fcm *FcmDelivery) *DefaultDeliveryService {
	return &DefaultDeliveryService{
		logger: logger.Named("delivery-service"),
		apns:   apns,
		fcm:    fcm,
	}
}

func (d DefaultDeliveryService) Send(ctx context.Context, req interfaces.SendRequest) error {
	d.logger.Info("Sending notification", zap.Any("req", req))
	var err error
	encodedMessage := base64.StdEncoding.EncodeToString(req.Message.Message)
	for _, installation := range req.Installations {
		if installation.DeliveryMechanism.Kind == interfaces.APNS && d.apns != nil {
			// TODO: Better error handling for cases where one message succeeds and another fails
			err = d.apns.Send(ctx, installation.DeliveryMechanism.Token, req.Message.GetContentTopic(), encodedMessage)
		}
		if installation.DeliveryMechanism.Kind == interfaces.FCM && d.fcm != nil {
			err = d.fcm.Send(ctx, installation.DeliveryMechanism.Token, req.Message.GetContentTopic(), encodedMessage)
		}
		if err != nil {
			break
		}
	}

	return err
}
