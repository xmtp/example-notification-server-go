package delivery

import (
	"context"
	"time"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/token"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	"go.uber.org/zap"
)

type DefaultDeliveryService struct {
	logger            *zap.Logger
	notificationTopic string
	apnsClient        *apns2.Client
}

func NewDeliveryService(logger *zap.Logger, opts options.ApnsOptions) (*DefaultDeliveryService, error) {
	client, err := getApnsClient([]byte(opts.P8Certificate), opts.KeyId, opts.TeamId)
	if err != nil {
		return nil, err
	}

	return &DefaultDeliveryService{
		logger:     logger.Named("delivery-service"),
		apnsClient: client,
	}, nil
}

func (d DefaultDeliveryService) Send(ctx context.Context, req interfaces.SendRequest) error {
	d.logger.Info("Sending notification", zap.Any("req", req))
	for _, installation := range req.Installations {
		if installation.DeliveryMechanism.Kind != interfaces.APNS {
			d.logger.Info("ignoring message. unsupported delivery mechanism")
			continue
		}
	}
	return nil
}

func (d DefaultDeliveryService) deliverApns(ctx context.Context, deviceToken, topic string) error {
	notification := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       "com.sideshow.Apns2",
		Payload:     []byte(`{"aps":{"alert":"Hello!"}}`),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := d.apnsClient.PushWithContext(ctx, notification)
	if res != nil {
		d.logger.Info("Sent notification", zap.String("apns_id", res.ApnsID))
	}
	return err
}

func getApnsClient(authKey []byte, keyId, teamId string) (*apns2.Client, error) {
	key, err := token.AuthKeyFromBytes(authKey)
	if err != nil {
		return nil, err
	}

	authToken := &token.Token{
		AuthKey: key,
		KeyID:   keyId,
		TeamID:  teamId,
	}

	return apns2.NewTokenClient(authToken), nil
}
