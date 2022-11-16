package api

import (
	"context"
	"log"

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
		logger:        logger,
		installations: installations,
		subscriptions: subscriptions,
	}
}

func (s *ApiServer) RegisterInstallation(
	ctx context.Context,
	req *connect.Request[proto.RegisterInstallationRequest],
) (*connect.Response[proto.RegisterInstallationResponse], error) {
	log.Println("Request headers: ", req.Header())
	res := connect.NewResponse(&proto.RegisterInstallationResponse{
		InstallationId: req.Msg.InstallationId,
	})

	return res, nil
}

func (s *ApiServer) DeleteInstallation(
	ctx context.Context,
	req *connect.Request[proto.DeleteInstallationRequest],
) (*connect.Response[emptypb.Empty], error) {
	log.Println("Request headers: ", req.Header())
	res := connect.NewResponse(&emptypb.Empty{})

	return res, nil
}

func (s *ApiServer) Subscribe(
	ctx context.Context,
	req *connect.Request[proto.SubscribeRequest],
) (*connect.Response[emptypb.Empty], error) {
	log.Println("Request headers: ", req.Header())
	res := connect.NewResponse(&emptypb.Empty{})

	return res, nil
}

func (s *ApiServer) Unsubscribe(
	ctx context.Context,
	req *connect.Request[proto.UnsubscribeRequest],
) (*connect.Response[emptypb.Empty], error) {
	log.Println("Request headers: ", req.Header())
	res := connect.NewResponse(&emptypb.Empty{})

	return res, nil
}
