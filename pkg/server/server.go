package server

import (
	"context"

	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	"github.com/xmtp/example-notification-server-go/pkg/xmtp"
	v1 "github.com/xmtp/proto/go/message_api/v1"
	"go.uber.org/zap"
)

type Server struct {
	ctx           context.Context
	logger        *zap.Logger
	xmtpClient    v1.MessageApiClient
	installations interfaces.Installations
	subscriptions interfaces.Subscriptions
	delivery      interfaces.Delivery
}

func New(ctx context.Context, opts options.Options, logger *zap.Logger, installations interfaces.Installations, subscriptions interfaces.Subscriptions, delivery interfaces.Delivery) (*Server, error) {
	client, err := xmtp.NewClient(ctx, opts.XmtpGrpcAddress)
	if err != nil {
		return nil, err
	}

	return &Server{
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
	return nil
}

func (s *Server) Stop() {
	s.logger.Info("Server stopped")
	return
}
