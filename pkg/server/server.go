package server

import (
	"context"
	"net/http"
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
	ctx           context.Context
	logger        *zap.Logger
	xmtpClient    v1.MessageApiClient
	installations interfaces.Installations
	subscriptions interfaces.Subscriptions
	delivery      interfaces.Delivery
	api           *api.ApiServer
	httpServer    *http.Server
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
		api:           api.NewApiServer(logger, installations, subscriptions),
		ctx:           ctx,
		logger:        logger,
		xmtpClient:    client,
		installations: installations,
		subscriptions: subscriptions,
		delivery:      delivery,
	}, nil
}

func (s *Server) Start() error {
	s.logger.Info("Server started")
	s.startApi()
	return nil
}

func (s *Server) startApi() {
	mux := http.NewServeMux()
	path, handler := protoconnect.NewNotificationsHandler(s.api)
	mux.Handle(path, handler)
	s.httpServer = &http.Server{
		Addr:    ":8080",
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil {
			s.logger.Fatal("server failed to start", zap.Error(err))
		}
	}()
}

func (s *Server) Stop() {
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Fatal("server failed to shutdown", zap.Error(err))
		}
	}

	s.logger.Info("Server stopped")
	return
}
