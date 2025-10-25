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
	l.logger.Info("🚀 Starting XMTP Listener",
		zap.String("xmtp_address", l.opts.GrpcAddress),
		zap.Bool("tls_enabled", l.opts.UseTls),
		zap.Int("num_workers", l.opts.NumWorkers),
		zap.Int("delivery_services_count", len(l.deliveryServices)),
	)

	// Log available delivery services
	for i := range l.deliveryServices {
		l.logger.Info("Delivery service registered",
			zap.Int("service_index", i+1),
		)
	}

	go l.startMessageListener()
	l.startMessageWorkers()

	l.logger.Info("✅ XMTP Listener started successfully",
		zap.Int("workers_spawned", l.opts.NumWorkers),
	)
}

func (l *Listener) Stop() {
	l.cancelFunc()
}

func (l *Listener) startMessageListener() {
	l.logger.Info("🔌 Connecting to XMTP network stream...",
		zap.String("address", l.opts.GrpcAddress),
	)
	var stream v1.MessageApi_SubscribeAllClient
	var err error
	sleepTime := STARTING_SLEEP_TIME
	for {
		stream, err = l.xmtpClient.SubscribeAll(l.ctx, &v1.SubscribeAllRequest{})
		if err != nil {
			l.logger.Error("❌ Error connecting to XMTP stream", zap.Error(err))
			l.logger.Info("⏳ Retrying connection...",
				zap.Duration("retry_delay", sleepTime),
			)
			time.Sleep(sleepTime)
			sleepTime = sleepTime * 2
			if err = l.refreshClient(); err != nil {
				l.logger.Error("error refreshing client", zap.Error(err))
			}
			continue
		}

		l.logger.Info("✅ Successfully connected to XMTP stream! Listening for messages...",
			zap.String("address", l.opts.GrpcAddress),
		)
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
					l.logger.Error("❌ Error reading from XMTP stream", zap.Error(err))
					l.logger.Info("⏳ Stream disconnected, attempting to reconnect...",
						zap.Duration("retry_delay", sleepTime),
					)
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
	l.logger.Info("👷 Starting message worker pool",
		zap.Int("worker_count", l.opts.NumWorkers),
	)

	for i := 0; i < l.opts.NumWorkers; i++ {
		workerID := i + 1
		l.logger.Debug("Starting worker",
			zap.Int("worker_id", workerID),
		)

		go func(id int) {
			var err error
			for msg := range l.messageChannel {
				err = l.processEnvelope(msg)
				if err != nil {
					l.logger.Error("error processing envelope",
						zap.Int("worker_id", id),
						zap.String("topic", msg.ContentTopic),
						zap.Error(err))
					continue
				}
			}
		}(workerID)
	}

	l.logger.Info("✅ All message workers started and ready")
}

func (l *Listener) processEnvelope(env *v1.Envelope) error {
	if !isV3Topic(env.ContentTopic) {
		l.logger.Debug("ignoring message", zap.String("topic", env.ContentTopic))
		return nil
	}

	// Generate unique message ID for tracking
	messageId := buildIdempotencyKey(env)

	l.logger.Info("🔍 Processing message for topic",
		zap.String("topic", env.ContentTopic),
		zap.String("message_id", messageId),
		zap.Uint64("timestamp_ns", env.TimestampNs),
		zap.Int("message_size", len(env.Message)),
	)

	subs, err := l.subscriptions.GetSubscriptions(l.ctx, env.ContentTopic, getThirtyDayPeriodsFromEpoch(env))
	if err != nil {
		return err
	}

	if len(subs) == 0 {
		l.logger.Debug("No subscriptions found for topic", zap.String("topic", env.ContentTopic))
		return nil
	}

	l.logger.Info("✅ Found subscriptions for topic",
		zap.String("topic", env.ContentTopic),
		zap.String("message_id", messageId),
		zap.Int("subscription_count", len(subs)),
	)

	installationIds := make([]string, len(subs))
	for i, sub := range subs {
		installationIds[i] = sub.InstallationId
		l.logger.Debug("Subscription details",
			zap.String("installation_id", sub.InstallationId),
			zap.String("topic", sub.Topic),
			zap.Bool("is_silent", sub.IsSilent),
			zap.Bool("is_active", sub.IsActive),
		)
	}

	installations, err := l.installations.GetInstallations(l.ctx, installationIds)
	if err != nil {
		return err
	}

	if len(installations) == 0 {
		l.logger.Info("No matching installations found for topic", zap.String("topic", env.ContentTopic))
		return nil
	}

	l.logger.Info("✅ Found installations for subscriptions",
		zap.Int("installation_count", len(installations)),
	)

	sendRequests := buildSendRequests(env, installations, subs)
	l.logger.Info("📦 Built send requests",
		zap.String("message_id", messageId),
		zap.Int("request_count", len(sendRequests)),
	)

	for idx, request := range sendRequests {
		shouldPushValue := "nil"
		if request.MessageContext.ShouldPush != nil {
			if *request.MessageContext.ShouldPush {
				shouldPushValue = "true"
			} else {
				shouldPushValue = "false"
			}
		}

		l.logger.Info("Processing send request",
			zap.String("message_id", request.IdempotencyKey),
			zap.Int("request_index", idx+1),
			zap.String("installation_id", request.Installation.Id),
			zap.String("delivery_mechanism", string(request.Installation.DeliveryMechanism.Kind)),
			zap.Bool("is_silent", request.Subscription.IsSilent),
			zap.String("should_push", shouldPushValue),
			zap.String("message_type", string(request.MessageContext.MessageType)),
		)

		if !l.shouldDeliver(request.MessageContext, request.Subscription) {
			l.logger.Info("⏭️ Skipping delivery of request",
				zap.String("message_id", request.IdempotencyKey),
				zap.String("installation_id", request.Installation.Id),
				zap.Any("message_context", request.MessageContext),
				zap.Bool("subscription_has_hmac_key", request.Subscription.HmacKey != nil),
			)
			continue
		}

		l.logger.Info("🚀 Attempting to deliver notification",
			zap.String("message_id", request.IdempotencyKey),
			zap.String("installation_id", request.Installation.Id),
			zap.String("topic", env.ContentTopic),
		)

		if err = l.deliver(request); err != nil {
			l.logger.Error("error delivering request", zap.Error(err), zap.String("content_topic", env.ContentTopic))
		}
	}
	return err
}

func (l *Listener) shouldDeliver(messageContext interfaces.MessageContext, subscription interfaces.Subscription) bool {
	shouldPushStr := "nil (not present)"
	if messageContext.ShouldPush != nil {
		if *messageContext.ShouldPush {
			shouldPushStr = "true"
		} else {
			shouldPushStr = "false"
		}
	}

	l.logger.Info("🔍 Checking delivery filters",
		zap.Bool("has_hmac_key", subscription.HmacKey != nil && len(subscription.HmacKey.Key) > 0),
		zap.String("should_push_value", shouldPushStr),
		zap.String("message_type", string(messageContext.MessageType)),
	)

	if subscription.HmacKey != nil && len(subscription.HmacKey.Key) > 0 {
		isSender := messageContext.IsSender(subscription.HmacKey.Key)
		l.logger.Info("HMAC sender check",
			zap.Bool("is_sender", isSender),
		)
		if isSender {
			l.logger.Info("❌ Skipping delivery: sender filter matched (user is sender)")
			return false
		}
	}

	if messageContext.ShouldPush != nil {
		shouldPush := messageContext.ShouldPush
		l.logger.Info("ShouldPush flag check",
			zap.Bool("should_push_value", *shouldPush),
		)
		if !*shouldPush {
			l.logger.Info("❌ Skipping delivery: ShouldPush flag is false")
		} else {
			l.logger.Info("✅ ShouldPush flag is true - will deliver")
		}
		return *shouldPush
	}

	l.logger.Info("✅ No filters blocking - will deliver (shouldPush not present, defaults to true)")
	return true
}

func (l *Listener) deliver(req interfaces.SendRequest) error {
	ctx, cancel := context.WithTimeout(l.ctx, DELIVERY_TIMEOUT)
	defer cancel()

	tokenPrefix := req.Installation.DeliveryMechanism.Token
	if len(tokenPrefix) > 20 {
		tokenPrefix = tokenPrefix[:20] + "..."
	}

	l.logger.Info("🔔 Delivery requested",
		zap.String("delivery_mechanism", string(req.Installation.DeliveryMechanism.Kind)),
		zap.String("token_prefix", tokenPrefix),
		zap.String("topic", req.Message.ContentTopic),
	)

	for _, service := range l.deliveryServices {
		if service.CanDeliver(req) && req.Message != nil {
			l.logger.Info("✅ Active subscription found. Sending notification via delivery service",
				zap.String("topic", req.Message.ContentTopic),
				zap.String("message_type", string(req.MessageContext.MessageType)),
				zap.String("delivery_mechanism", string(req.Installation.DeliveryMechanism.Kind)),
			)
			return service.Send(ctx, req)
		}
	}
	l.logger.Warn("⚠️ No delivery service matches request",
		zap.String("delivery_mechanism", string(req.Installation.DeliveryMechanism.Kind)),
		zap.Int("available_services", len(l.deliveryServices)),
	)
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

func isV3Topic(topic string) bool {
	if strings.HasPrefix(topic, "/xmtp/mls/1/g-") || strings.HasPrefix(topic, "/xmtp/mls/1/w-") {
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
