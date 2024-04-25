package xmtp

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"strings"
	"time"

	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	v1 "github.com/xmtp/example-notification-server-go/pkg/proto/message_api/v1"
	"go.uber.org/zap"
)

const STARTING_SLEEP_TIME = 100 * time.Millisecond
const DELIVERY_TIMEOUT = 15 * time.Second

type Listener struct {
	logger           *zap.Logger
	ctx              context.Context
	cancelFunc       func()
	xmtpClient       v1.MessageApiClient
	opts             options.XmtpOptions
	messageChannel   chan *v1.Envelope
	installations    interfaces.Installations
	deliveryServices []interfaces.Delivery
	subscriptions    interfaces.Subscriptions
	clientVersion    string
	appVersion       string
}

func NewListener(
	ctx context.Context,
	logger *zap.Logger,
	opts options.XmtpOptions,
	installations interfaces.Installations,
	subscriptions interfaces.Subscriptions,
	deliveryServices []interfaces.Delivery,
	clientVersion string,
	appVersion string,
) (*Listener, error) {
	client, err := NewClient(ctx, opts.GrpcAddress, opts.UseTls, clientVersion, appVersion)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	return &Listener{
		ctx:              ctx,
		cancelFunc:       cancel,
		logger:           logger.Named("xmtp-listener"),
		xmtpClient:       client,
		opts:             opts,
		messageChannel:   make(chan *v1.Envelope, 100),
		installations:    installations,
		deliveryServices: deliveryServices,
		subscriptions:    subscriptions,
		clientVersion:    clientVersion,
		appVersion:       appVersion,
	}, nil
}

func (l *Listener) Start() {
	go l.startMessageListener()
	l.startMessageWorkers()
}

func (l *Listener) Stop() {
	l.cancelFunc()
}

func (l *Listener) startMessageListener() {
	l.logger.Info("starting message listener")
	var stream v1.MessageApi_SubscribeAllClient
	var err error
	sleepTime := STARTING_SLEEP_TIME
	for {
		stream, err = l.xmtpClient.SubscribeAll(l.ctx, &v1.SubscribeAllRequest{})
		if err != nil {
			l.logger.Error("error connecting to stream", zap.Error(err))
			time.Sleep(sleepTime)
			sleepTime = sleepTime * 2
			if err = l.refreshClient(); err != nil {
				l.logger.Error("error refreshing client", zap.Error(err))
			}
			continue
		}
	streamLoop:
		for {
			select {
			case <-l.ctx.Done():
				close(l.messageChannel)
				return
			default:
				msg, err := stream.Recv()
				if err == io.EOF {
					l.logger.Info("stream closed")
					break streamLoop
				}

				if err != nil {
					l.logger.Error("error reading from stream", zap.Error(err))
					// Wait 100ms to avoid hammering the API and getting rate limited
					time.Sleep(sleepTime)
					sleepTime = sleepTime * 2
					if err = l.refreshClient(); err != nil {
						l.logger.Error("error refreshing client", zap.Error(err))
					}
					break streamLoop
				}

				if msg != nil {
					// Reset the sleep time on first successful message
					sleepTime = STARTING_SLEEP_TIME
					l.messageChannel <- msg
				}
			}
		}
	}
}

func (l *Listener) startMessageWorkers() {
	for i := 0; i < l.opts.NumWorkers; i++ {
		go func() {
			var err error
			for msg := range l.messageChannel {
				err = l.processEnvelope(msg)
				if err != nil {
					l.logger.Error("error processing envelope", zap.String("topic", msg.ContentTopic), zap.Error(err))
					continue
				}
				// l.logger.Info("processed a message", zap.String("topic", msg.ContentTopic))
			}
		}()
	}
}

func (l *Listener) processEnvelope(env *v1.Envelope) error {
	if shouldIgnoreTopic(env.ContentTopic) {
		l.logger.Debug("ignoring message", zap.String("topic", env.ContentTopic))
		return nil
	}

	subs, err := l.subscriptions.GetSubscriptions(l.ctx, env.ContentTopic, getThirtyDayPeriodsFromEpoch(env))
	if err != nil {
		return err
	}

	if len(subs) == 0 {
		return nil
	}

	installationIds := make([]string, len(subs))
	for i, sub := range subs {
		installationIds[i] = sub.InstallationId
	}

	installations, err := l.installations.GetInstallations(l.ctx, installationIds)
	if err != nil {
		return err
	}

	if len(installations) == 0 {
		l.logger.Info("No matching installations found for topic", zap.String("topic", env.ContentTopic))
		return nil
	}

	sendRequests := buildSendRequests(env, installations, subs)
	for _, request := range sendRequests {
		if !l.shouldDeliver(request.MessageContext, request.Subscription) {
			l.logger.Info("Skipping delivery of request",
				zap.Any("message_context", request.MessageContext),
				zap.Bool("subscription_has_hmac_key", request.Subscription.HmacKey != nil),
			)
			continue
		}
		if err = l.deliver(request); err != nil {
			l.logger.Error("error delivering request", zap.Error(err), zap.String("content_topic", env.ContentTopic))
		}
	}
	return err
}

func (l *Listener) shouldDeliver(messageContext interfaces.MessageContext, subscription interfaces.Subscription) bool {
	if subscription.HmacKey != nil && len(subscription.HmacKey.Key) > 0 {
		isSender := messageContext.IsSender(subscription.HmacKey.Key)
		if isSender {
			return false
		}
	}
	if messageContext.ShouldPush != nil {
		shouldPush := messageContext.ShouldPush
		return *shouldPush
	}
	return true
}

func (l *Listener) deliver(req interfaces.SendRequest) error {
	ctx, cancel := context.WithTimeout(l.ctx, DELIVERY_TIMEOUT)
	defer cancel()
	for _, service := range l.deliveryServices {
		if service.CanDeliver(req) && req.Message != nil {
			l.logger.Info("active subscription found. sending message",
				zap.String("topic", req.Message.ContentTopic),
				zap.String("message_type", string(req.MessageContext.MessageType)),
			)
			return service.Send(ctx, req)
		}
	}
	l.logger.Info("No delivery service matches request", zap.String("delivery_mechanism", string(req.Installation.DeliveryMechanism.Kind)))
	return nil
}

func (l *Listener) refreshClient() error {
	client, err := NewClient(l.ctx, l.opts.GrpcAddress, l.opts.UseTls, l.clientVersion, l.appVersion)
	if err != nil {
		return err
	}
	l.xmtpClient = client

	return nil
}

func shouldIgnoreTopic(topic string) bool {
	if strings.HasPrefix(topic, "/xmtp/0/contact-") || strings.HasPrefix(topic, "/xmtp/0/privatestore-") {
		return true
	}
	return false
}

func buildIdempotencyKey(env *v1.Envelope) string {
	h := sha1.New()
	h.Write([]byte(env.ContentTopic))
	h.Write(env.Message)
	return hex.EncodeToString(h.Sum(nil))
}

func buildSendRequests(envelope *v1.Envelope, installations []interfaces.Installation, subscriptions []interfaces.Subscription) []interfaces.SendRequest {
	idempotencyKey := buildIdempotencyKey(envelope)
	messageContext := getContext(envelope)
	out := []interfaces.SendRequest{}
	installationMap := make(map[string]interfaces.Installation)
	for _, installation := range installations {
		installationMap[installation.Id] = installation
	}

	for _, subscription := range subscriptions {
		if installation, exists := installationMap[subscription.InstallationId]; exists {
			out = append(out, interfaces.SendRequest{
				IdempotencyKey: idempotencyKey,
				Message:        envelope,
				MessageContext: messageContext,
				Installation:   installation,
				Subscription:   subscription,
			})
		}
	}

	return out
}
