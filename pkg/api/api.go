package api

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ApiServer struct {
	logger        *zap.Logger
	installations interfaces.Installations
	subscriptions interfaces.Subscriptions
}

func NewApiServer(logger *zap.Logger, installations interfaces.Installations, subscriptions interfaces.Subscriptions) *ApiServer {
	return &ApiServer{
		logger:        logger.Named("api"),
		installations: installations,
		subscriptions: subscriptions,
	}
}

func (s *ApiServer) RegisterInstallation(
	ctx context.Context,
	req *connect.Request[proto.RegisterInstallationRequest],
) (*connect.Response[proto.RegisterInstallationResponse], error) {
	s.logger.Info("RegisterInstallation", zap.Any("headers", req.Header()), zap.Any("req", req))
	res := connect.NewResponse(&proto.RegisterInstallationResponse{
		InstallationId: req.Msg.InstallationId,
	})

	return res, nil
}

func (s *ApiServer) DeleteInstallation(
	ctx context.Context,
	req *connect.Request[proto.DeleteInstallationRequest],
) (*connect.Response[emptypb.Empty], error) {
	s.logger.Info("DeleteInstallation", zap.Any("headers", req.Header()), zap.Any("req", req))
	res := connect.NewResponse(&emptypb.Empty{})

	return res, nil
}

func (s *ApiServer) Subscribe(
	ctx context.Context,
	req *connect.Request[proto.SubscribeRequest],
) (*connect.Response[emptypb.Empty], error) {
	s.logger.Info("Subscribe", zap.Any("headers", req.Header()), zap.Any("req", req))
	res := connect.NewResponse(&emptypb.Empty{})

	return res, nil
}

func (s *ApiServer) Unsubscribe(
	ctx context.Context,
	req *connect.Request[proto.UnsubscribeRequest],
) (*connect.Response[emptypb.Empty], error) {
	s.logger.Info("Unsubscribe", zap.Any("headers", req.Header()), zap.Any("req", req))
	res := connect.NewResponse(&emptypb.Empty{})

	return res, nil
}
