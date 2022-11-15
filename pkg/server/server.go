package server

import (
	"context"

	"github.com/xmtp/example-notification-server-go/pkg/options"
	"github.com/xmtp/example-notification-server-go/pkg/xmtp"
	v1 "github.com/xmtp/proto/go/message_api/v1"
	"go.uber.org/zap"
)

type Server struct {
	ctx           context.Context
	cancel        context.CancelFunc
	xmtpClient    v1.MessageApiClient
	installations InstallationService
	subscriptions SubscriptionService
	delivery      DeliveryService
}

func New(logger *zap.Logger, opts options.Options, installations InstallationService, subscriptions SubscriptionService, delivery DeliveryService) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())
	client, err := xmtp.NewClient(ctx, opts.XmtpAddress)
	if err != nil {
		cancel()
		return nil, err
	}

	return &Server{
		ctx:           ctx,
		cancel:        cancel,
		xmtpClient:    client,
		installations: installations,
		subscriptions: subscriptions,
		delivery:      delivery,
	}, nil
}

func (s *Server) RunUntilShutdown() {
	return
}
