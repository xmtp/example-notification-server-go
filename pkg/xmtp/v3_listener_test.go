package xmtp

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/xmtp/example-notification-server-go/mocks"
	"github.com/xmtp/example-notification-server-go/pkg/installations"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	"github.com/xmtp/example-notification-server-go/pkg/subscriptions"
	topics "github.com/xmtp/example-notification-server-go/pkg/topics"
	"github.com/xmtp/example-notification-server-go/pkg/testutils"
	v1 "github.com/xmtp/xmtpd/pkg/proto/message_api/v1"
)

const (
	XMTP_ADDRESS      = "localhost:25556"
	INSTALLATION_ID   = "test_installation"
	INSTALLATION_ID_2 = "test_installation_2"
	TEST_TOPIC        = "/xmtp/mls/1/w-abcdef0123456789/proto"
	DELIVERY_TOKEN    = "test_token"
)

func buildTestListener(t *testing.T, deliveryService interfaces.Delivery) *Listener {
	t.Helper()
	logger := testutils.TestLogger(t)
	ctx, cancel := context.WithCancel(t.Context())
	opts := options.XmtpOptions{ListenerEnabled: true, GrpcAddress: XMTP_ADDRESS, UseTls: false, NumWorkers: 5}
	db := testutils.CreateTestDb(t)
	installations := installations.NewInstallationsService(logger, db)
	subscriptions := subscriptions.NewSubscriptionsService(logger, db)

	l, err := NewListener(ctx, logger, opts, installations, subscriptions, []interfaces.Delivery{deliveryService}, "test", "test")
	if err != nil {
		require.NoError(t, err)
	}
	l.Start()

	t.Cleanup(func() {
		cancel()
		l.Stop()
	})

	return l
}

func injectMessage(listener *Listener, topic string, message []byte) {
	listener.messageChannel <- &v1.Envelope{
		ContentTopic: topic,
		Message:      message,
		TimestampNs:  uint64(time.Now().UnixNano()),
	}
}

func subscribeToTopic(t *testing.T, l *Listener, installationId, topicStr string, isSilent bool) {
	_, err := l.installations.Register(t.Context(), interfaces.Installation{
		Id: installationId,
		DeliveryMechanism: interfaces.DeliveryMechanism{
			Kind:  interfaces.APNS,
			Token: DELIVERY_TOKEN,
		},
	})
	require.NoError(t, err)

	parsed, err := topics.ParseV3Topic(topicStr)
	require.NoError(t, err)

	err = l.subscriptions.SubscribeWithMetadata(t.Context(), installationId, []interfaces.SubscriptionInput{{Topic: parsed, IsSilent: isSilent}})
	require.NoError(t, err)
}

func Test_BasicDelivery(t *testing.T) {
	mockDeliveryService, sendCount := testutils.MockDeliveryWithSendCounter(t)
	l := buildTestListener(t, mockDeliveryService)

	subscribeToTopic(t, l, INSTALLATION_ID, TEST_TOPIC, false)
	injectMessage(l, TEST_TOPIC, []byte("test"))
	testutils.RequireEventuallySendCount(t, sendCount, 1)

	mockDeliveryService.AssertCalled(t, "CanDeliver", mock.Anything)
	mockDeliveryService.AssertNumberOfCalls(t, "Send", 1)

	sendReqs := testutils.GetSendRequests(mockDeliveryService)
	require.Len(t, sendReqs, 1)
	require.Equal(t, INSTALLATION_ID, sendReqs[0].Installation.Id)
	require.Equal(t, TEST_TOPIC, sendReqs[0].Topic)
	require.Equal(t, topics.V3Welcome, sendReqs[0].MessageContext.MessageType)
}

func Test_MultipleDeliveries(t *testing.T) {
	mockDeliveryService := mocks.NewDelivery(t)
	l := buildTestListener(t, mockDeliveryService)

	mockDeliveryService.On("CanDeliver", mock.Anything).Return(true)
	var sendCount atomic.Int32
	mockDeliveryService.On("Send", mock.Anything, mock.Anything).
		Run(func(mock.Arguments) {
			sendCount.Add(1)
		}).
		Once().
		Return(errors.New("failed"))
	mockDeliveryService.On("Send", mock.Anything, mock.Anything).
		Run(func(mock.Arguments) {
			sendCount.Add(1)
		}).
		Once().
		Return(nil)

	subscribeToTopic(t, l, INSTALLATION_ID, TEST_TOPIC, false)
	subscribeToTopic(t, l, INSTALLATION_ID_2, TEST_TOPIC, false)

	injectMessage(l, TEST_TOPIC, []byte("test"))
	testutils.RequireEventuallySendCount(t, &sendCount, 2)

	mockDeliveryService.AssertCalled(t, "CanDeliver", mock.Anything)
	mockDeliveryService.AssertNumberOfCalls(t, "Send", 2)

	sendReqs := testutils.GetSendRequests(mockDeliveryService)
	require.Len(t, sendReqs, 2)
	require.ElementsMatch(t, []string{INSTALLATION_ID, INSTALLATION_ID_2}, []string{
		sendReqs[0].Installation.Id,
		sendReqs[1].Installation.Id,
	})
	require.Equal(t, TEST_TOPIC, sendReqs[0].Topic)
	require.Equal(t, TEST_TOPIC, sendReqs[1].Topic)
}
