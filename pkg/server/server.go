package server

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/xmtp/example-notification-server-go/pkg/api"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	"github.com/xmtp/example-notification-server-go/pkg/proto/protoconnect"
	"github.com/xmtp/example-notification-server-go/pkg/xmtp"
	v1 "github.com/xmtp/proto/go/message_api/v1"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Server struct {
	ctx            context.Context
	logger         *zap.Logger
	opts           options.Options
	messageChannel chan *v1.Envelope
	xmtpClient     v1.MessageApiClient
	installations  interfaces.Installations
	subscriptions  interfaces.Subscriptions
	delivery       interfaces.Delivery
	api            *api.ApiServer
	httpServer     *http.Server
}

func New(
	ctx context.Context,
	opts options.Options,
	logger *zap.Logger,
	installations interfaces.Installations,
	subscriptions interfaces.Subscriptions,
	delivery interfaces.Delivery,
) (*Server, error) {
	client, err := xmtp.NewClient(ctx, opts.XmtpGrpcAddress)
	if err != nil {
		return nil, err
	}

	return &Server{
		api:            api.NewApiServer(logger, installations, subscriptions),
		ctx:            ctx,
		messageChannel: make(chan *v1.Envelope, 100),
		opts:           opts,
		logger:         logger,
		xmtpClient:     client,
		installations:  installations,
		subscriptions:  subscriptions,
		delivery:       delivery,
	}, nil
}

func (s *Server) Start() error {
	s.logger.Info("server starting")

	if s.opts.Api.Enabled {
		s.startApi()
	}

	if s.opts.Worker.Enabled {
		s.startMessageWorkers()
		go s.startMessageListener()
	}

	return nil
}

func (s *Server) startApi() {
	mux := http.NewServeMux()
	path, handler := protoconnect.NewNotificationsHandler(s.api)
	mux.Handle(path, handler)
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.opts.Api.Port),
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	s.logger.Info("api server started", zap.String("address", s.httpServer.Addr), zap.String("path", path))

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil {
			s.logger.Info("api server stopped", zap.Error(err))
		}
	}()
}

func (s *Server) startMessageListener() {
	s.logger.Info("starting message listener")
	var stream v1.MessageApi_SubscribeAllClient
	var err error
	for {
		stream, err = s.xmtpClient.SubscribeAll(s.ctx, &v1.SubscribeAllRequest{})
		if err != nil {
			s.logger.Error("error connecting to stream", zap.Error(err))
			// sleep for a few seconds before retrying
			time.Sleep(3 * time.Second)
			continue
		}
		for {
			select {
			case <-s.ctx.Done():
				close(s.messageChannel)
				return
			default:
				msg, err := stream.Recv()
				if err == io.EOF {
					s.logger.Info("stream closed")
					break
				}

				if err != nil {
					s.logger.Error("error reading from stream", zap.Error(err))
					break
				}

				if msg != nil {
					s.messageChannel <- msg
				}
			}
		}
	}
}

func (s *Server) startMessageWorkers() {
	for i := 0; i < s.opts.Worker.NumWorkers; i++ {
		go func() {
			var err error
			for msg := range s.messageChannel {
				err = s.processEnvelope(msg)
				if err != nil {
					s.logger.Error("error processing envelope", zap.String("topic", msg.ContentTopic), zap.Error(err))
					continue
				}
				s.logger.Info("processed a message", zap.String("topic", msg.ContentTopic))
			}
			s.logger.Info("shutting down worker")
		}()
	}
}

func (s *Server) processEnvelope(env *v1.Envelope) error {
	if shouldIgnoreTopic(env.ContentTopic) {
		s.logger.Info("ignoring message", zap.String("topic", env.ContentTopic))
		return nil
	}

	subs, err := s.subscriptions.GetSubscriptions(s.ctx, env.ContentTopic)
	if err != nil {
		return err
	}

	installationIds := make([]string, len(subs))
	for i, sub := range subs {
		installationIds[i] = sub.InstallationId
	}

	installations, err := s.installations.GetInstallations(s.ctx, installationIds)
	if err != nil {
		return err
	}

	if len(installations) == 0 {
		return nil
	}

	return s.delivery.Send(
		s.ctx,
		interfaces.SendRequest{
			Installations:  installations,
			Message:        env,
			IdempotencyKey: buildIdempotencyKey(env),
		},
	)
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

func (s *Server) Stop() {
	s.logger.Info("server shutting down")
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Fatal("server failed to shutdown", zap.Error(err))
		}
	}

	s.logger.Info("server stopped")
	return
}
