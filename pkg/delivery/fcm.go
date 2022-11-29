package delivery

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/pkg/errors"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

type FcmDelivery struct {
	logger *zap.Logger
	client *messaging.Client
}

func NewFcmDelivery(ctx context.Context, logger *zap.Logger, opts options.FcmOptions) (*FcmDelivery, error) {
	creds := option.WithCredentialsJSON([]byte(opts.CredentialsJson))
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: opts.ProjectId,
	}, creds)

	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize firebase app")
	}

	// Use the auth method to validate the credentials
	_, err = app.Auth(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "firebase credentials failed to validate")
	}

	messaging, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	return &FcmDelivery{
		logger: logger,
		client: messaging,
	}, nil
}

func (f *FcmDelivery) Send(ctx context.Context, token, topic, message string) error {
	data := map[string]string{
		"topic":            topic,
		"encryptedMessage": message,
	}

	_, err := f.client.Send(ctx, &messaging.Message{
		Token: token,
		Data:  data,
		Android: &messaging.AndroidConfig{
			Data:     data,
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-push-type": "background",
				"apns-priority":  "5",
			},
			Payload: &messaging.APNSPayload{
				CustomData: map[string]interface {
				}{
					"topic":            topic,
					"encryptedMessage": message,
				},
				Aps: &messaging.Aps{
					ContentAvailable: true,
				},
			},
		},
	})

	return err
}
