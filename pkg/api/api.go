package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/xmtp/example-notification-server-go/pkg/delivery"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	v1 "github.com/xmtp/example-notification-server-go/pkg/proto/message_api/v1"
	proto "github.com/xmtp/example-notification-server-go/pkg/proto/notifications/v1"
	"github.com/xmtp/example-notification-server-go/pkg/proto/notifications/v1/notificationsv1connect"
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
	expoOpts      options.ExpoOptions
}

func NewApiServer(logger *zap.Logger, opts options.ApiOptions, expoOpts options.ExpoOptions, installations interfaces.Installations, subscriptions interfaces.Subscriptions) *ApiServer {
	return &ApiServer{
		logger:        logger.Named("api"),
		installations: installations,
		subscriptions: subscriptions,
		port:          opts.Port,
		expoOpts:      expoOpts,
	}
}

func (s *ApiServer) Start() {
	mux := http.NewServeMux()
	path, handler := notificationsv1connect.NewNotificationsHandler(s)
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
}

func (s *ApiServer) RegisterInstallation(
	ctx context.Context,
	req *connect.Request[proto.RegisterInstallationRequest],
) (*connect.Response[proto.RegisterInstallationResponse], error) {
	s.logger.Info("RegisterInstallation", zap.Any("req", req))

	mechanism := convertDeliveryMechanism(req.Msg.DeliveryMechanism)
	if mechanism == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("missing delivery mechanism"))
	}
	s.logger.Info("got mechanism", zap.Any("mechanism", mechanism))
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

	// Send test notification for Expo registrations
	if mechanism.Kind == interfaces.EXPO {
		s.logger.Info("Sending test Expo notification")
		go s.sendTestNotification(context.Background(), req.Msg.InstallationId, *mechanism)
	}

	s.logger.Info("sending response", zap.Any("result", result))
	return connect.NewResponse(&proto.RegisterInstallationResponse{
		InstallationId: req.Msg.InstallationId,
		ValidUntil:     uint64(result.ValidUntil.UnixMilli()),
	}), nil
}

func (s *ApiServer) DeleteInstallation(
	ctx context.Context,
	req *connect.Request[proto.DeleteInstallationRequest],
) (*connect.Response[emptypb.Empty], error) {
	s.logger.Info("DeleteInstallation", zap.Any("req", req))

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
	s.logger.Info("Subscribe", zap.Any("req", req))

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
	s.logger.Info("Unsubscribe", zap.Any("req", req))

	err := s.subscriptions.Unsubscribe(ctx, req.Msg.InstallationId, req.Msg.Topics)
	if err != nil {
		s.logger.Error("error unsubscribing", zap.Error(err))
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *ApiServer) SubscribeWithMetadata(ctx context.Context, req *connect.Request[proto.SubscribeWithMetadataRequest]) (*connect.Response[emptypb.Empty], error) {
	log := s.logger.With(zap.String("method", "subscribeWithMetadata"))
	log.Info("Subscribing")
	inputs := buildSubscriptionInputs(req.Msg.Subscriptions)
	err := s.subscriptions.SubscribeWithMetadata(ctx, req.Msg.InstallationId, inputs)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func buildSubscriptionInputs(subs []*proto.Subscription) []interfaces.SubscriptionInput {
	out := make([]interfaces.SubscriptionInput, len(subs))
	for idx, sub := range subs {
		out[idx] = interfaces.SubscriptionInput{
			Topic:    sub.Topic,
			IsSilent: sub.IsSilent,
			HmacKeys: buildHmacKeys(sub.HmacKeys),
		}
	}
	return out
}

func buildHmacKeys(protoKeys []*proto.Subscription_HmacKey) []interfaces.HmacKey {
	out := make([]interfaces.HmacKey, len(protoKeys))
	for idx, key := range protoKeys {
		out[idx] = interfaces.HmacKey{
			ThirtyDayPeriodsSinceEpoch: int(key.ThirtyDayPeriodsSinceEpoch),
			Key:                        key.Key,
		}
	}
	return out
}

func convertDeliveryMechanism(mechanism *proto.DeliveryMechanism) *interfaces.DeliveryMechanism {
	if mechanism == nil {
		return nil
	}
	apnsToken := mechanism.GetApnsDeviceToken()
	fcmToken := mechanism.GetFirebaseDeviceToken()
	customToken := mechanism.GetCustomToken()

	if apnsToken != "" {
		return &interfaces.DeliveryMechanism{Kind: interfaces.APNS, Token: apnsToken}
	} else if fcmToken != "" {
		return &interfaces.DeliveryMechanism{Kind: interfaces.FCM, Token: fcmToken}
	} else if customToken != "" {
		// For Expo tokens (ExponentPushToken[...])
		return &interfaces.DeliveryMechanism{Kind: interfaces.EXPO, Token: customToken}
	}
	return nil
}

func (s *ApiServer) sendTestNotification(ctx context.Context, installationId string, mechanism interfaces.DeliveryMechanism) {
	s.logger.Info("Preparing test notification",
		zap.String("installationId", installationId),
		zap.String("token", mechanism.Token))

	// Create Expo delivery instance
	expoDelivery := delivery.NewExpoDelivery(s.logger, s.expoOpts)

	// Create a test message
	testMessage := []byte("Welcome to XMTP notifications!")

	// Create a dummy send request for the test notification
	testRequest := interfaces.SendRequest{
		Installation: interfaces.Installation{
			Id:                installationId,
			DeliveryMechanism: mechanism,
		},
		Subscription: interfaces.Subscription{
			IsSilent: false,
		},
		MessageContext: interfaces.MessageContext{},
		Message: &v1.Envelope{
			ContentTopic: "/xmtp/test",
			Message:      testMessage,
		},
	}

	// Send the test notification
	err := expoDelivery.Send(ctx, testRequest)
	if err != nil {
		s.logger.Error("Failed to send test notification", zap.Error(err))
	} else {
		s.logger.Info("Test notification sent successfully!", zap.String("installationId", installationId))
	}
}
