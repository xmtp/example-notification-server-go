package delivery

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
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
		log.Fatalf("error initializing app: %v\n", err)
	}
	if err != nil {
		return nil, err
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
	title := "New message from XMTP"
	body := "Open app to read"
	data := map[string]string{
		"topic":            topic,
		"encryptedMessage": message,
	}

	_, err := f.client.Send(ctx, &messaging.Message{
		Token: token,
		Data:  data,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Android: &messaging.AndroidConfig{
			Notification: &messaging.AndroidNotification{
				Title: title,
				Body:  body,
			},
			Data: data,
		},
		Webpush: &messaging.WebpushConfig{},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				CustomData: map[string]interface {
				}{
					"topic":            topic,
					"encryptedMessage": message,
				},
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: title,
						Body:  body,
					},
					MutableContent: true,
				},
			},
		},
	})

	return err
}
