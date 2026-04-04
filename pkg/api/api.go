package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	proto "github.com/xmtp/example-notification-server-go/pkg/proto/notifications/v1"
	"github.com/xmtp/example-notification-server-go/pkg/proto/notifications/v1/notificationsv1connect"
	topicutil "github.com/xmtp/example-notification-server-go/pkg/topics"
	"github.com/xmtp/xmtpd/pkg/topic"
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
	listener      net.Listener
}

func NewApiServer(logger *zap.Logger, opts options.ApiOptions, installations interfaces.Installations, subscriptions interfaces.Subscriptions) *ApiServer {
	return &ApiServer{
		logger:        logger.Named("api"),
		installations: installations,
		subscriptions: subscriptions,
		port:          opts.Port,
	}
}

func (s *ApiServer) SetListener(listener net.Listener) error {
	if s.httpServer != nil {
		return errors.New("api server already started")
	}

	s.listener = listener
	if tcpAddr, ok := listener.Addr().(*net.TCPAddr); ok {
		s.port = tcpAddr.Port
	}

	return nil
}

func (s *ApiServer) Start() {
	mux := http.NewServeMux()
	path, handler := notificationsv1connect.NewNotificationsHandler(s)
	mux.Handle(path, handler)

	addr := fmt.Sprintf(":%d", s.port)
	if s.listener != nil {
		addr = s.listener.Addr().String()
	}
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	s.logger.Info("api server started", zap.String("address", s.httpServer.Addr), zap.Int("port", s.port), zap.String("path", path))

	go func() {
		var err error
		if s.listener != nil {
			err = s.httpServer.Serve(s.listener)
		} else {
			err = s.httpServer.ListenAndServe()
		}
		if err != nil {
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

	topics, err := normalizeTopics(req.Msg.Topics, req.Msg.TopicsBytes)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	err = s.subscriptions.Subscribe(ctx, req.Msg.InstallationId, topics)
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

	topics, err := normalizeTopics(req.Msg.Topics, req.Msg.TopicsBytes)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	err = s.subscriptions.Unsubscribe(ctx, req.Msg.InstallationId, topics)
	if err != nil {
		s.logger.Error("error unsubscribing", zap.Error(err))
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *ApiServer) SubscribeWithMetadata(ctx context.Context, req *connect.Request[proto.SubscribeWithMetadataRequest]) (*connect.Response[emptypb.Empty], error) {
	log := s.logger.With(zap.String("method", "subscribeWithMetadata"))
	log.Info("Subscribing")
	inputs, err := normalizeSubscriptionInputs(req.Msg.Subscriptions)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	err = s.subscriptions.SubscribeWithMetadata(ctx, req.Msg.InstallationId, inputs)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func normalizeTopics(stringTopics []string, bytesTopics [][]byte) ([]*topic.Topic, error) {
	seen := make(map[string]struct{})
	var result []*topic.Topic

	for _, s := range stringTopics {
		t, err := topicutil.ParseV3Topic(s)
		if err != nil {
			return nil, fmt.Errorf("invalid topic %q: %w", s, err)
		}
		key := string(t.Bytes())
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			result = append(result, t)
		}
	}

	for _, b := range bytesTopics {
		t, err := topic.ParseTopic(b)
		if err != nil {
			return nil, fmt.Errorf("invalid binary topic: %w", err)
		}
		key := string(t.Bytes())
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			result = append(result, t)
		}
	}

	return result, nil
}

func normalizeSubscriptionInputs(subs []*proto.Subscription) ([]interfaces.SubscriptionInput, error) {
	out := make([]interfaces.SubscriptionInput, len(subs))
	for idx, sub := range subs {
		var t *topic.Topic
		var err error
		if len(sub.TopicBytes) > 0 {
			t, err = topic.ParseTopic(sub.TopicBytes)
			if err != nil {
				return nil, fmt.Errorf("invalid binary topic at index %d: %w", idx, err)
			}
		} else if sub.Topic != "" {
			t, err = topicutil.ParseV3Topic(sub.Topic)
			if err != nil {
				return nil, fmt.Errorf("invalid topic %q at index %d: %w", sub.Topic, idx, err)
			}
		} else {
			return nil, fmt.Errorf("subscription at index %d has no topic", idx)
		}
		out[idx] = interfaces.SubscriptionInput{
			Topic:    t,
			IsSilent: sub.IsSilent,
			HmacKeys: buildHmacKeys(sub.HmacKeys),
		}
	}
	return out, nil
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
	if apnsToken != "" {
		return &interfaces.DeliveryMechanism{Kind: interfaces.APNS, Token: apnsToken}
	} else {
		return &interfaces.DeliveryMechanism{Kind: interfaces.FCM, Token: fcmToken}
	}
}
