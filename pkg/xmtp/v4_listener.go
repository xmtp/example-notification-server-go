package xmtp

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"time"

	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	"github.com/xmtp/example-notification-server-go/pkg/topics"
	"github.com/xmtp/xmtpd/pkg/envelopes"
	mlsV1 "github.com/xmtp/xmtpd/pkg/proto/mls/api/v1"
	envelopesProto "github.com/xmtp/xmtpd/pkg/proto/xmtpv4/envelopes"
	notificationApi "github.com/xmtp/xmtpd/pkg/proto/xmtpv4/message_api"
	"github.com/xmtp/xmtpd/pkg/topic"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type V4Listener struct {
	dispatcher      deliveryDispatcher
	logger          *zap.Logger
	ctx             context.Context
	cancelFunc      func()
	v4Client        notificationApi.NotificationApiClient
	opts            options.XmtpOptions
	envelopeChannel chan *envelopesProto.OriginatorEnvelope
	installations   interfaces.Installations
	subscriptions   interfaces.Subscriptions
	clientVersion   string
	appVersion      string
}

func NewV4Listener(
	ctx context.Context,
	logger *zap.Logger,
	opts options.XmtpOptions,
	installations interfaces.Installations,
	subscriptions interfaces.Subscriptions,
	deliveryServices []interfaces.Delivery,
	clientVersion string,
	appVersion string,
) (*V4Listener, error) {
	client, err := NewV4Client(ctx, opts.GrpcAddress, opts.UseTls, clientVersion, appVersion)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	namedLogger := logger.Named("xmtp-v4-listener")

	return &V4Listener{
		ctx:             ctx,
		cancelFunc:      cancel,
		logger:          namedLogger,
		v4Client:        client,
		opts:            opts,
		envelopeChannel: make(chan *envelopesProto.OriginatorEnvelope, 100),
		installations:   installations,
		subscriptions:   subscriptions,
		clientVersion:   clientVersion,
		appVersion:      appVersion,
		dispatcher: deliveryDispatcher{
			logger:           namedLogger,
			ctx:              ctx,
			deliveryServices: deliveryServices,
		},
	}, nil
}

func (l *V4Listener) Start() {
	go l.startEnvelopeListener()
	l.startEnvelopeWorkers()
}

func (l *V4Listener) Stop() {
	l.cancelFunc()
}

func (l *V4Listener) startEnvelopeListener() {
	l.logger.Info("starting V4 envelope listener")
	sleepTime := STARTING_SLEEP_TIME
	for {
		stream, err := l.v4Client.SubscribeAllEnvelopes(l.ctx, &notificationApi.SubscribeAllEnvelopesRequest{})
		if err != nil {
			l.logger.Error("error connecting to V4 stream", zap.Error(err))
			time.Sleep(sleepTime)
			sleepTime = sleepTime * 2
			if err = l.refreshV4Client(); err != nil {
				l.logger.Error("error refreshing V4 client", zap.Error(err))
			}
			continue
		}
	streamLoop:
		for {
			select {
			case <-l.ctx.Done():
				close(l.envelopeChannel)
				return
			default:
				resp, err := stream.Recv()
				if err == io.EOF {
					l.logger.Info("V4 stream closed")
					break streamLoop
				}

				if err != nil {
					l.logger.Error("error reading from V4 stream", zap.Error(err))
					// Wait to avoid hammering the API and getting rate limited
					time.Sleep(sleepTime)
					sleepTime = sleepTime * 2
					if err = l.refreshV4Client(); err != nil {
						l.logger.Error("error refreshing V4 client", zap.Error(err))
					}
					break streamLoop
				}

				if resp != nil {
					// Reset the sleep time on first successful message
					sleepTime = STARTING_SLEEP_TIME
					for _, env := range resp.GetEnvelopes() {
						l.envelopeChannel <- env
					}
				}
			}
		}
	}
}

func (l *V4Listener) startEnvelopeWorkers() {
	for i := 0; i < l.opts.NumWorkers; i++ {
		go func() {
			for env := range l.envelopeChannel {
				if err := l.processOriginatorEnvelope(env); err != nil {
					l.logger.Error("error processing originator envelope", zap.Error(err))
				}
			}
		}()
	}
}

func (l *V4Listener) processOriginatorEnvelope(env *envelopesProto.OriginatorEnvelope) error {
	origEnv, err := envelopes.NewOriginatorEnvelope(env)
	if err != nil {
		l.logger.Info("skipping envelope: failed to parse originator envelope", zap.Error(err))
		return nil
	}

	clientEnvProto := origEnv.UnsignedOriginatorEnvelope.PayerEnvelope.ClientEnvelope.Proto()
	if clientEnvProto == nil {
		l.logger.Info("skipping envelope: client envelope proto is nil")
		return nil
	}

	aad := clientEnvProto.GetAad()
	if aad == nil {
		l.logger.Info("skipping envelope: AAD is nil")
		return nil
	}

	targetTopicBytes := aad.GetTargetTopic()
	if len(targetTopicBytes) == 0 {
		l.logger.Info("skipping envelope: target topic is empty")
		return nil
	}

	targetTopic, err := topic.ParseTopic(targetTopicBytes)
	if err != nil {
		l.logger.Info("skipping envelope: failed to parse target topic", zap.Error(err))
		return nil
	}

	thirtyDayPeriod := int(origEnv.OriginatorNs() / 1_000_000_000 / 60 / 60 / 24 / 30)
	subs, err := l.subscriptions.GetSubscriptions(l.ctx, targetTopic, thirtyDayPeriod)
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

	insts, err := l.installations.GetInstallations(l.ctx, installationIds)
	if err != nil {
		return err
	}

	if len(insts) == 0 {
		return nil
	}

	idempotencyKey := buildV4IdempotencyKey(env)
	installationMap := make(map[string]interfaces.Installation, len(insts))
	for _, inst := range insts {
		installationMap[inst.Id] = inst
	}

	for _, sub := range subs {
		inst, exists := installationMap[sub.InstallationId]
		if !exists {
			continue
		}

		req, skip := l.buildV4SendRequest(env, origEnv, clientEnvProto, targetTopic, idempotencyKey, inst, sub)
		if skip {
			continue
		}

		if !l.dispatcher.shouldDeliver(req.MessageContext, req.Subscription) {
			l.logger.Info("skipping delivery of V4 request",
				zap.Any("message_context", req.MessageContext),
				zap.Bool("subscription_has_hmac_key", req.Subscription.HmacKey != nil),
			)
			continue
		}

		if err = l.dispatcher.deliver(req); err != nil {
			l.logger.Error("error delivering V4 request", zap.Error(err))
		}
	}

	return nil
}

// buildV4SendRequest constructs a SendRequest for the given installation.
// Returns (request, true) if the request should be skipped, (request, false) otherwise.
func (l *V4Listener) buildV4SendRequest(
	env *envelopesProto.OriginatorEnvelope,
	origEnv *envelopes.OriginatorEnvelope,
	clientEnvProto *envelopesProto.ClientEnvelope,
	targetTopic *topic.Topic,
	idempotencyKey string,
	inst interfaces.Installation,
	sub interfaces.Subscription,
) (interfaces.SendRequest, bool) {
	if inst.PayloadFormat == interfaces.PayloadFormatV4 {
		// V4 format: deliver raw OriginatorEnvelope bytes
		envBytes, err := proto.Marshal(env)
		if err != nil {
			l.logger.Error("failed to marshal originator envelope for V4 delivery", zap.Error(err))
			return interfaces.SendRequest{}, true
		}

		var messageContext interfaces.MessageContext
		switch payload := clientEnvProto.GetPayload().(type) {
		case *envelopesProto.ClientEnvelope_GroupMessage:
			v1Input := payload.GroupMessage.GetV1()
			messageContext = buildGroupMessageContext(v1Input)
		case *envelopesProto.ClientEnvelope_WelcomeMessage:
			messageContext = interfaces.MessageContext{MessageType: topics.V3Welcome}
		default:
			messageContext = interfaces.MessageContext{MessageType: topics.Unknown}
		}

		return interfaces.SendRequest{
			IdempotencyKey:   idempotencyKey,
			Topic:            topics.TopicToBase64(targetTopic),
			EncryptedMessage: envBytes,
			PayloadFormat:    interfaces.PayloadFormatV4,
			MessageContext:   messageContext,
			Installation:     inst,
			Subscription:     sub,
		}, false
	}

	// V3 format (PayloadFormatV3 or PayloadFormatUnspecified)
	logTopic := topics.TopicToLegacy(targetTopic)
	switch payload := clientEnvProto.GetPayload().(type) {
	case *envelopesProto.ClientEnvelope_GroupMessage:
		clientEnvelope := origEnv.UnsignedOriginatorEnvelope.PayerEnvelope.ClientEnvelope
		if !clientEnvelope.TopicMatchesPayload() {
			l.logger.Error("group message target topic does not match payload")
			return interfaces.SendRequest{}, true
		}
		v1Input := payload.GroupMessage.GetV1()
		encryptedMsg, err := convertGroupMessageToV3(v1Input, origEnv, targetTopic)
		if err != nil {
			logV4ConversionIssue(l.logger, clientEnvProto, err, logTopic)
			return interfaces.SendRequest{}, true
		}
		messageContext := buildGroupMessageContext(v1Input)
		return interfaces.SendRequest{
			IdempotencyKey:   idempotencyKey,
			Topic:            logTopic,
			EncryptedMessage: encryptedMsg,
			PayloadFormat:    interfaces.PayloadFormatV3,
			MessageContext:   messageContext,
			Installation:     inst,
			Subscription:     sub,
		}, false

	case *envelopesProto.ClientEnvelope_WelcomeMessage:
		clientEnvelope := origEnv.UnsignedOriginatorEnvelope.PayerEnvelope.ClientEnvelope
		if !clientEnvelope.TopicMatchesPayload() {
			l.logger.Error("welcome message target topic does not match payload")
			return interfaces.SendRequest{}, true
		}
		var encryptedMsg []byte
		var err error
		if v1Input := payload.WelcomeMessage.GetV1(); v1Input != nil {
			encryptedMsg, err = convertWelcomeMessageToV3(v1Input, origEnv)
		} else if wpInput := payload.WelcomeMessage.GetWelcomePointer(); wpInput != nil {
			encryptedMsg, err = convertWelcomePointerToV3(wpInput, origEnv)
		} else {
			l.logger.Warn("welcome message has unknown version, skipping V3 conversion")
			return interfaces.SendRequest{}, true
		}
		if err != nil {
			logV4ConversionIssue(l.logger, clientEnvProto, err, logTopic)
			return interfaces.SendRequest{}, true
		}
		return interfaces.SendRequest{
			IdempotencyKey:   idempotencyKey,
			Topic:            logTopic,
			EncryptedMessage: encryptedMsg,
			PayloadFormat:    interfaces.PayloadFormatV3,
			MessageContext:   interfaces.MessageContext{MessageType: topics.V3Welcome},
			Installation:     inst,
			Subscription:     sub,
		}, false

	default:
		// Non-convertible payload: skip V3 installations
		logV4ConversionIssue(l.logger, clientEnvProto, nil, logTopic)
		return interfaces.SendRequest{}, true
	}
}

func buildGroupMessageContext(v1Input *mlsV1.GroupMessageInput_V1) interfaces.MessageContext {
	if v1Input == nil {
		return interfaces.MessageContext{MessageType: topics.V3Conversation}
	}
	shouldPush := v1Input.ShouldPush
	hmacInputs := cloneBytes(v1Input.Data)
	senderHmac := cloneBytes(v1Input.SenderHmac)
	mc := interfaces.MessageContext{
		MessageType: topics.V3Conversation,
		ShouldPush:  &shouldPush,
		HmacInputs:  &hmacInputs,
	}
	if len(senderHmac) > 0 {
		mc.SenderHmac = &senderHmac
	}
	return mc
}

func logV4ConversionIssue(logger *zap.Logger, clientEnvProto *envelopesProto.ClientEnvelope, err error, logTopic string) {
	fields := []zap.Field{zap.String("topic", logTopic), zap.Error(err)}
	if clientEnvProto == nil {
		logger.Warn("v4 payload cannot be converted to v3", fields...)
		return
	}
	switch clientEnvProto.GetPayload().(type) {
	case *envelopesProto.ClientEnvelope_GroupMessage, *envelopesProto.ClientEnvelope_WelcomeMessage:
		logger.Error("v4 group/welcome payload conversion to v3 failed", fields...)
	default:
		logger.Warn("v4 payload cannot be converted to v3", fields...)
	}
}

func buildV4IdempotencyKey(env *envelopesProto.OriginatorEnvelope) string {
	h := sha1.New()
	b, err := proto.Marshal(env)
	if err == nil {
		h.Write(b)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (l *V4Listener) refreshV4Client() error {
	client, err := NewV4Client(l.ctx, l.opts.GrpcAddress, l.opts.UseTls, l.clientVersion, l.appVersion)
	if err != nil {
		return err
	}
	l.v4Client = client
	return nil
}
