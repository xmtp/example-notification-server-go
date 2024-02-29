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

type Listener struct {
	logger         *zap.Logger
	ctx            context.Context
	cancelFunc     func()
	xmtpClient     v1.MessageApiClient
	opts           options.XmtpOptions
	messageChannel chan *v1.Envelope
	installations  interfaces.Installations
	delivery       interfaces.Delivery
	subscriptions  interfaces.Subscriptions
	clientVersion  string
	appVersion     string
}

func NewListener(
	ctx context.Context,
	logger *zap.Logger,
	opts options.XmtpOptions,
	installations interfaces.Installations,
	subscriptions interfaces.Subscriptions,
	delivery interfaces.Delivery,
	clientVersion string,
	appVersion string,
) (*Listener, error) {
	client, err := NewClient(ctx, opts.GrpcAddress, opts.UseTls, clientVersion, appVersion)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	return &Listener{
		ctx:            ctx,
		cancelFunc:     cancel,
		logger:         logger.Named("xmtp-listener"),
		xmtpClient:     client,
		opts:           opts,
		messageChannel: make(chan *v1.Envelope, 100),
		installations:  installations,
		delivery:       delivery,
		subscriptions:  subscriptions,
		clientVersion:  clientVersion,
		appVersion:     appVersion,
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
	for {
		stream, err = l.xmtpClient.SubscribeAll(l.ctx, &v1.SubscribeAllRequest{})
		if err != nil {
			l.logger.Error("error connecting to stream", zap.Error(err))
			// sleep for a few seconds before retrying
			time.Sleep(100 * time.Millisecond)
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
					time.Sleep(100 * time.Millisecond)
					if err = l.refreshClient(); err != nil {
						l.logger.Error("error refreshing client", zap.Error(err))
					}
					break streamLoop
				}

				if msg != nil {
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
				l.logger.Info("processed a message", zap.String("topic", msg.ContentTopic))
			}
		}()
	}
}

func (l *Listener) processEnvelope(env *v1.Envelope) error {
	if shouldIgnoreTopic(env.ContentTopic) {
		l.logger.Info("ignoring message", zap.String("topic", env.ContentTopic))
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

	l.logger.Info("active subscription found. sending message", zap.String("topic", env.ContentTopic))

	return l.delivery.Send(
		l.ctx,
		interfaces.SendRequest{
			Installations:  installations,
			Message:        env,
			IdempotencyKey: buildIdempotencyKey(env),
		},
	)
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
