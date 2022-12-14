package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	"github.com/xmtp/example-notification-server-go/pkg/proto"
	"github.com/xmtp/example-notification-server-go/pkg/proto/protoconnect"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ApiServer struct {
	logger        *zap.Logger
	installations interfaces.Installations
	subscriptions interfaces.Subscriptions
	httpServer    *http.Server
	port          int
}

func NewApiServer(logger *zap.Logger, opts options.ApiOptions, installations interfaces.Installations, subscriptions interfaces.Subscriptions) *ApiServer {
	return &ApiServer{
		logger:        logger.Named("api"),
		installations: installations,
		subscriptions: subscriptions,
		port:          opts.Port,
	}
}

func (s *ApiServer) Start() {
	mux := http.NewServeMux()
	path, handler := protoconnect.NewNotificationsHandler(s)
	mux.Handle(path, handler)
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	s.logger.Info("api server started", zap.String("address", s.httpServer.Addr), zap.Int("port", s.port), zap.String("path", path))

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil {
			s.logger.Info("api server stopped", zap.Error(err))
		}
	}()
}

func (s *ApiServer) Stop() {
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

func (s *ApiServer) RegisterInstallation(
	ctx context.Context,
	req *connect.Request[proto.RegisterInstallationRequest],
) (*connect.Response[proto.RegisterInstallationResponse], error) {
	s.logger.Info("RegisterInstallation", zap.Any("headers", req.Header()), zap.Any("req", req))

	mechanism := convertDeliveryMechanism(req.Msg.DeliveryMechanism)
	if mechanism == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("missing delivery mechanism"))
	}
	result, err := s.installations.Register(
		ctx,
		interfaces.Installation{
			Id:                req.Msg.InstallationId,
			DeliveryMechanism: *mechanism,
		},
	)

	if err != nil {
		s.logger.Error("error registering installation", zap.Error(err))
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&proto.RegisterInstallationResponse{
		InstallationId: req.Msg.InstallationId,
		ValidUntil:     uint64(result.ValidUntil.UnixMilli()),
	}), nil
}

func (s *ApiServer) DeleteInstallation(
	ctx context.Context,
	req *connect.Request[proto.DeleteInstallationRequest],
) (*connect.Response[emptypb.Empty], error) {
	s.logger.Info("DeleteInstallation", zap.Any("headers", req.Header()), zap.Any("req", req))

	err := s.installations.Delete(ctx, req.Msg.InstallationId)
	if err != nil {
		s.logger.Error("error deleting installation", zap.Error(err))
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *ApiServer) Subscribe(
	ctx context.Context,
	req *connect.Request[proto.SubscribeRequest],
) (*connect.Response[emptypb.Empty], error) {
	s.logger.Info("Subscribe", zap.Any("headers", req.Header()), zap.Any("req", req))

	err := s.subscriptions.Subscribe(ctx, req.Msg.InstallationId, req.Msg.Topics)
	if err != nil {
		s.logger.Error("error subscribing", zap.Error(err))
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *ApiServer) Unsubscribe(
	ctx context.Context,
	req *connect.Request[proto.UnsubscribeRequest],
) (*connect.Response[emptypb.Empty], error) {
	s.logger.Info("Unsubscribe", zap.Any("headers", req.Header()), zap.Any("req", req))

	err := s.subscriptions.Unsubscribe(ctx, req.Msg.InstallationId, req.Msg.Topics)
	if err != nil {
		s.logger.Error("error unsubscribing", zap.Error(err))
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func convertDeliveryMechanism(mechanism *proto.DeliveryMechanism) *interfaces.DeliveryMechanism {
	if mechanism == nil {
		return nil
	}
	apnsToken := mechanism.GetApnsDeviceToken()
	fcmToken := mechanism.GetFirebaseDeviceToken()
	if apnsToken != "" {
		return &interfaces.DeliveryMechanism{Kind: interfaces.APNS, Token: apnsToken}
	} else {
		return &interfaces.DeliveryMechanism{Kind: interfaces.FCM, Token: fcmToken}
	}
}
