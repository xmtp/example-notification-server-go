package xmtp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/xmtp/example-notification-server-go/mocks"
	"github.com/xmtp/example-notification-server-go/pkg/installations"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/logging"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	v1 "github.com/xmtp/example-notification-server-go/pkg/proto/message_api/v1"
	"github.com/xmtp/example-notification-server-go/pkg/subscriptions"
	"github.com/xmtp/example-notification-server-go/test"
)

const (
	XMTP_ADDRESS      = "localhost:25556"
	INSTALLATION_ID   = "test_installation"
	INSTALLATION_ID_2 = "test_installation_2"
	TEST_TOPIC        = "/xmtp/0/test_topic/proto"
	DELIVERY_TOKEN    = "test_token"
)

func buildTestListener(t *testing.T, deliveryService interfaces.Delivery) (*Listener, func()) {
	logger := logging.CreateLogger("console", "info")
	ctx, cancel := context.WithCancel(context.Background())
	opts := options.XmtpOptions{ListenerEnabled: true, GrpcAddress: XMTP_ADDRESS, UseTls: false, NumWorkers: 5}
	db, cleanup := test.CreateTestDb()
	installations := installations.NewInstallationsService(logger, db)
	subscriptions := subscriptions.NewSubscriptionsService(logger, db)

	l, err := NewListener(ctx, logger, opts, installations, subscriptions, []interfaces.Delivery{deliveryService}, "test", "test")
	if err != nil {
		require.NoError(t, err)
	}
	l.Start()

	return l, func() {
		cancel()
		l.Stop()
		cleanup()
	}
}

func sendMessage(t *testing.T, listener *Listener, topic string, message []byte) {
	_, err := listener.xmtpClient.Publish(context.Background(), &v1.PublishRequest{
		Envelopes: []*v1.Envelope{
			{
				ContentTopic: topic,
				Message:      message,
				TimestampNs:  uint64(time.Now().UnixNano()),
			},
		},
	})
	require.NoError(t, err)
}

func subscribeToTopic(t *testing.T, l *Listener, installationId, topic string, isSilent bool) {
	_, err := l.installations.Register(context.Background(), interfaces.Installation{
		Id: installationId,
		DeliveryMechanism: interfaces.DeliveryMechanism{
			Kind:  interfaces.APNS,
			Token: DELIVERY_TOKEN,
		},
	})
	require.NoError(t, err)

	err = l.subscriptions.SubscribeWithMetadata(context.Background(), installationId, []interfaces.SubscriptionInput{{Topic: topic, IsSilent: isSilent}})
	require.NoError(t, err)
}

func Test_BasicDelivery(t *testing.T) {
	mockDeliveryService := mocks.NewDelivery(t)
	l, cleanup := buildTestListener(t, mockDeliveryService)
	defer cleanup()

	mockDeliveryService.On("CanDeliver", mock.Anything).Return(true)
	mockDeliveryService.On("Send", mock.Anything, mock.Anything).Return(nil)

	subscribeToTopic(t, l, INSTALLATION_ID, TEST_TOPIC, false)
	sendMessage(t, l, TEST_TOPIC, []byte("test"))
	time.Sleep(2 * time.Second)

	mockDeliveryService.AssertCalled(t, "CanDeliver", mock.Anything)
	mockDeliveryService.AssertCalled(t, "Send", mock.Anything, mock.Anything)
	mockDeliveryService.AssertNumberOfCalls(t, "Send", 1)
}

func Test_MultipleDeliveries(t *testing.T) {
	mockDeliveryService := mocks.NewDelivery(t)
	l, cleanup := buildTestListener(t, mockDeliveryService)
	defer cleanup()

	mockDeliveryService.On("CanDeliver", mock.Anything).Return(true)
	mockDeliveryService.On("Send", mock.Anything, mock.Anything).Once().Return(errors.New("failed"))
	mockDeliveryService.On("Send", mock.Anything, mock.Anything).Once().Return(nil)

	subscribeToTopic(t, l, INSTALLATION_ID, TEST_TOPIC, false)
	subscribeToTopic(t, l, INSTALLATION_ID_2, TEST_TOPIC, false)

	sendMessage(t, l, TEST_TOPIC, []byte("test"))
	time.Sleep(2 * time.Second)

	mockDeliveryService.AssertCalled(t, "CanDeliver", mock.Anything)
	mockDeliveryService.AssertCalled(t, "Send", mock.Anything, mock.Anything)
	mockDeliveryService.AssertNumberOfCalls(t, "Send", 2)
}
