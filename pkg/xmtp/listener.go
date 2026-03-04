package xmtp

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	v1 "github.com/xmtp/example-notification-server-go/pkg/proto/message_api/v1"
	"github.com/xmtp/example-notification-server-go/pkg/proto/xmtpv4/envelopes"
)

const STARTING_SLEEP_TIME = 100 * time.Millisecond
const DELIVERY_TIMEOUT = 15 * time.Second

type Listener struct {
	logger     *zap.Logger
	ctx        context.Context
	cancelFunc func()
	opts       options.XmtpOptions

	client    SubscriberClient
	envelopes chan any

	installations    interfaces.Installations
	deliveryServices []interfaces.Delivery
	subscriptions    interfaces.Subscriptions

	clientVersion string
	appVersion    string
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

	ctx, cancel := context.WithCancel(ctx)

	conn, err := newConn(opts.GrpcAddress, opts.UseTls, clientVersion, appVersion)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("could not initialize GRPC client: %w", err)
	}

	logger.Info("creating xmtp listener", zap.Bool("d14n", opts.D14N))

	client := newSubscriberClient(conn, UseV3Client(!opts.D14N))

	listener := &Listener{
		ctx:              ctx,
		cancelFunc:       cancel,
		logger:           logger.Named("xmtp-listener"),
		opts:             opts,
		client:           client,
		envelopes:        make(chan any),
		installations:    installations,
		deliveryServices: deliveryServices,
		subscriptions:    subscriptions,
		clientVersion:    clientVersion,
		appVersion:       appVersion,
	}

	return listener, nil
}

func (l *Listener) Start() {

	// NOTE: Tentative support for cursor is added but it seems clunky
	// to specify a key-value map via CLI flags.
	// Since it's dubious if it's necessary, mechanism for specifying it is
	// not added right now.
	go l.startEnvelopeListener(nil)

	l.startMessageWorkers()
}

func (l *Listener) Stop() {
	l.cancelFunc()
}

func (l *Listener) startEnvelopeListener(cursor map[uint32]uint64) {

	l.logger.Info("starting message listener")

	var (
		stream SubscriberStream
		err    error
	)

	sleepTime := STARTING_SLEEP_TIME
	for {
		stream, err = l.client.Subscribe(l.ctx, cursor)
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
				close(l.envelopes)
				return
			default:

				msg, err := stream.Receive()
				if err != nil && errors.Is(err, io.EOF) {
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
				}

				// Range over envelopes so they get distributed to the worker pool evenly.
				// Only one, either v3 or v4 will be populated, but we can just range over both.
				for _, env := range msg.V3 {
					l.envelopes <- env
				}

				for _, env := range msg.V4 {
					l.envelopes <- env
				}
			}
		}
	}
}

func (l *Listener) startMessageWorkers() {
	for i := 0; i < l.opts.NumWorkers; i++ {
		go func() {
			for msg := range l.envelopes {

				switch env := msg.(type) {

				case *v1.Envelope:
					err := l.processV3Envelope(env)
					if err != nil {
						l.logger.Error("could not process envelope",
							zap.String("topic", env.ContentTopic),
							zap.Error(err),
						)
					}

				case *envelopes.OriginatorEnvelope:

					err := l.processV4Envelope(env)
					if err != nil {
						l.logger.Error("could not process v4 envelope",
							zap.Error(err))
					}
				}
			}
		}()
	}
}

func shouldDeliver(messageContext interfaces.MessageContext, subscription interfaces.Subscription) bool {
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

	if req.Empty() {
		return errors.New("empty message scheduled for delivery, skipping")
	}

	ctx, cancel := context.WithTimeout(l.ctx, DELIVERY_TIMEOUT)
	defer cancel()

	// TODO: Fix - we range accross multiple delivery services, but only the first one will be picked up.
	for _, service := range l.deliveryServices {

		if !service.CanDeliver(req) {
			continue
		}

		l.logger.Info("active subscription found. sending message",
			zap.String("topic", req.GetTopic()),
			zap.String("message_type", string(req.MessageContext.MessageType)),
		)

		return service.Send(ctx, req)
	}

	l.logger.Info("No delivery service matches request", zap.String("delivery_mechanism", string(req.Installation.DeliveryMechanism.Kind)))
	return nil
}

func (l *Listener) refreshClient() error {

	conn, err := newConn(l.opts.GrpcAddress, l.opts.UseTls, l.clientVersion, l.appVersion)
	if err != nil {
		return fmt.Errorf("could not refresh GRPC client: %w", err)
	}

	l.client = newSubscriberClient(conn, UseV3Client(!l.opts.D14N))

	return nil
}

// isV3Topic returns true if topic is one we care about - group message or welcome message.
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

func buildIdempotencyKeyV4(info messageV4Info) string {
	h := sha1.New()
	h.Write([]byte(info.context.Topic))

	// NOTE: Idempotency data is initialized to group message .GetV1().GetData().
	// HmacInputs are set to the same thing, so there's duplication there.
	h.Write(*info.idempotencyData)
	return hex.EncodeToString(h.Sum(nil))
}

func buildSendRequestV4(env *envelopes.OriginatorEnvelope, info messageV4Info, installations []interfaces.Installation, subscriptions []interfaces.Subscription) []interfaces.SendRequest {

	var (
		idempotencyKey = buildIdempotencyKeyV4(info)
		out            []interfaces.SendRequest
	)

	installationMap := make(map[string]interfaces.Installation)
	for _, installation := range installations {
		installationMap[installation.Id] = installation
	}

	for _, sub := range subscriptions {

		inst, ok := installationMap[sub.InstallationId]
		if !ok {
			continue
		}

		out = append(out, interfaces.SendRequest{
			IdempotencyKey: idempotencyKey,
			MessageV4:      env,
			MessageContext: info.context,
			Installation:   inst,
			Subscription:   sub,
		})
	}

	return out

}
