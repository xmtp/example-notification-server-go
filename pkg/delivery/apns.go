package delivery

import (
	"context"
	"errors"
	"os"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	"go.uber.org/zap"
)

type ApnsDelivery struct {
	logger     *zap.Logger
	apnsClient *apns2.Client
	opts       options.ApnsOptions
}

func NewApnsDelivery(logger *zap.Logger, opts options.ApnsOptions) (*ApnsDelivery, error) {
	var bytes []byte
	var err error

	if opts.P8Certificate == "" {
		bytes, err = os.ReadFile(opts.P8CertificateFilePath)

		if err != nil {
			return nil, err
		}
	} else {
		bytes = []byte(opts.P8Certificate)
	}

	client, err := getApnsClient(bytes, opts.KeyId, opts.TeamId)
	if err != nil {
		return nil, err
	}

	switch opts.Mode {
	case "production":
		client.Production()
	case "development":
		client.Development()
	default:
		return nil, errors.New("invalid APNS mode")
	}

	return &ApnsDelivery{
		logger:     logger.Named("delivery-service"),
		apnsClient: client,
		opts:       opts,
	}, nil
}

func (a ApnsDelivery) CanDeliver(req interfaces.SendRequest) bool {
	return req.Installation.DeliveryMechanism.Kind == interfaces.APNS
}

func (a ApnsDelivery) Send(ctx context.Context, req interfaces.SendRequest) error {
	if req.Message == nil {
		return errors.New("missing message")
	}

	notification := a.buildNotification(req.Subscription.IsSilent,
		req.Installation.DeliveryMechanism.Token,
		req.Message.ContentTopic,
		string(req.MessageContext.MessageType),
		req.Message.Message,
	)

	res, err := a.apnsClient.PushWithContext(ctx, notification)
	if res != nil {
		a.logger.Info(
			"Sent notification",
			zap.String("apns_id", res.ApnsID),
			zap.Int("status_code", res.StatusCode),
			zap.String("reason", res.Reason),
		)
	}

	return err
}

func (a ApnsDelivery) buildNotification(isSilent bool, token string, contentTopic string, messageKind string, messageBytes []byte) *apns2.Notification {
	notificationPayload := payload.NewPayload().
		Custom("topic", contentTopic).
		Custom("encryptedMessage", messageBytes).
		Custom("messageKind", messageKind)

	if isSilent {
		notificationPayload = notificationPayload.ContentAvailable()
	} else {
		notificationPayload = notificationPayload.
			Alert("New message from XMTP").
			MutableContent()
	}

	return &apns2.Notification{
		DeviceToken: token,
		Topic:       a.opts.Topic,
		Payload:     notificationPayload,
	}
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
