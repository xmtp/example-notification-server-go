package xmtp

import (
	"context"
	"io"
	"time"

	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	envelopesProto "github.com/xmtp/xmtpd/pkg/proto/xmtpv4/envelopes"
	notificationApi "github.com/xmtp/xmtpd/pkg/proto/xmtpv4/message_api"
	"go.uber.org/zap"
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
	return nil
}

func (l *V4Listener) refreshV4Client() error {
	client, err := NewV4Client(l.ctx, l.opts.GrpcAddress, l.opts.UseTls, l.clientVersion, l.appVersion)
	if err != nil {
		return err
	}
	l.v4Client = client
	return nil
}
