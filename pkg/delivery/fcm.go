package delivery

import (
	"fmt"
	"context"
	"encoding/base64"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/pkg/errors"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	"github.com/xmtp/example-notification-server-go/pkg/topics"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

type FcmDelivery struct {
	logger *zap.Logger
	client *messaging.Client
	env string
}

func NewFcmDelivery(ctx context.Context, logger *zap.Logger, opts options.FcmOptions, env string) (*FcmDelivery, error) {
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
		env: env
	}, nil
}

func (f FcmDelivery) CanDeliver(req interfaces.SendRequest) bool {
	return req.Installation.DeliveryMechanism.Kind == interfaces.FCM && req.Installation.DeliveryMechanism.Token != ""
}

func (f FcmDelivery) Send(ctx context.Context, req interfaces.SendRequest) error {
	if req.Message == nil {
		return errors.New("missing message")
	}

	message := base64.StdEncoding.EncodeToString(req.Message.Message)
	topic := req.Message.ContentTopic
	data := map[string]string{
		"topic":            topic,
		"encryptedMessage": message,
		"messageType":      string(req.MessageContext.MessageType),
	}

	prefix := ""
	if f.env != "prod" {
		prefix = fmt.Sprintf("%s.", env)
	}

	link := fmt.Sprintf("https://%shopscotch.trade/chat", prefix)
	if req.MessageContext.MessageType == topics.V2Invite || req.MessageContext.MessageType == topics.V1Intro {
		link = fmt.Sprintf("https://%shopscotch.trade/chat/invites", prefix)
	}

	webpushHeaders := map[string]string{}
	webpushHeaders["Urgency"] = "high"

	apnsHeaders := map[string]string{}
	androidPriority := "high"

	if req.Subscription.IsSilent {
		apnsHeaders["apns-push-type"] = "background"
		apnsHeaders["apns-priority"] = "5"
		androidPriority = "normal"
		webpushHeaders["Urgency"] = "normal"
	}

	_, err := f.client.Send(ctx, &messaging.Message{
		Token: req.Installation.DeliveryMechanism.Token,
		Data:  data,
		Android: &messaging.AndroidConfig{
			Data:     data,
			Priority: androidPriority,
		},
		Webpush: &messaging.WebpushConfig{
			Data: data,
			Headers: webpushHeaders,
			FCMOptions: &messaging.WebpushFCMOptions{
				Link: link
			},
		},
		APNS: &messaging.APNSConfig{
			Headers: apnsHeaders,
			Payload: &messaging.APNSPayload{
				CustomData: map[string]interface {
				}{
					"topic":            topic,
					"encryptedMessage": message,
					"messageType":      string(req.MessageContext.MessageType),
				},
				Aps: &messaging.Aps{
					ContentAvailable: req.Subscription.IsSilent,
					MutableContent:   !req.Subscription.IsSilent,
				},
			},
		},
	})

	return err
}
