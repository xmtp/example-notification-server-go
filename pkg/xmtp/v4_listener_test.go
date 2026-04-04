package xmtp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/xmtp/example-notification-server-go/mocks"
	"github.com/xmtp/example-notification-server-go/pkg/installations"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/logging"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	"github.com/xmtp/example-notification-server-go/pkg/subscriptions"
	"github.com/xmtp/example-notification-server-go/test"
	mlsV1 "github.com/xmtp/xmtpd/pkg/proto/mls/api/v1"
	envelopesProto "github.com/xmtp/xmtpd/pkg/proto/xmtpv4/envelopes"
	testEnvelopes "github.com/xmtp/xmtpd/pkg/testutils/envelopes"
	"github.com/xmtp/xmtpd/pkg/topic"
	"google.golang.org/protobuf/proto"
)

func TestV4Listener_NewAndStop(t *testing.T) {
	logger := logging.CreateLogger("console", "info")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db := test.CreateTestDb(t)
	instSvc := installations.NewInstallationsService(logger, db)
	subsSvc := subscriptions.NewSubscriptionsService(logger, db)
	mockDelivery := mocks.NewDelivery(t)

	l, err := NewV4Listener(ctx, logger, options.XmtpOptions{
		ListenerEnabled: true,
		GrpcAddress:     "localhost:25556",
		NumWorkers:      5,
	}, instSvc, subsSvc, []interfaces.Delivery{mockDelivery}, "test", "test")
	require.NoError(t, err)
	require.NotNil(t, l)
	l.Stop()
}

// getSendRequests returns all SendRequest arguments captured by a mock Delivery's Send calls.
func getSendRequests(mockDelivery *mocks.Delivery) []interfaces.SendRequest {
	var reqs []interfaces.SendRequest
	for _, call := range mockDelivery.Calls {
		if call.Method == "Send" {
			reqs = append(reqs, call.Arguments.Get(1).(interfaces.SendRequest))
		}
	}
	return reqs
}

// buildV4TestListener creates a V4Listener with real DB services and mock delivery.
func buildV4TestListener(t *testing.T, deliveryService interfaces.Delivery) *V4Listener {
	t.Helper()
	logger := logging.CreateLogger("console", "info")
	ctx := context.Background()
	db := test.CreateTestDb(t)
	instSvc := installations.NewInstallationsService(logger, db)
	subsSvc := subscriptions.NewSubscriptionsService(logger, db)

	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	namedLogger := logger.Named("xmtp-v4-listener")
	return &V4Listener{
		ctx:             ctx,
		cancelFunc:      cancel,
		logger:          namedLogger,
		opts:            options.XmtpOptions{NumWorkers: 5},
		envelopeChannel: make(chan *envelopesProto.OriginatorEnvelope, 100),
		installations:   instSvc,
		subscriptions:   subsSvc,
		dispatcher: deliveryDispatcher{
			logger:           namedLogger,
			ctx:              ctx,
			deliveryServices: []interfaces.Delivery{deliveryService},
		},
	}
}

// registerV4Installation registers an installation with the given payload format.
func registerV4Installation(t *testing.T, l *V4Listener, installationID string, format interfaces.PayloadFormat) {
	t.Helper()
	_, err := l.installations.Register(context.Background(), interfaces.Installation{
		Id: installationID,
		DeliveryMechanism: interfaces.DeliveryMechanism{
			Kind:  interfaces.APNS,
			Token: "test-token",
		},
		PayloadFormat: format,
	})
	require.NoError(t, err)
}

// subscribeV4ToTopic subscribes the given installation to a topic.
func subscribeV4ToTopic(t *testing.T, l *V4Listener, installationID string, tp *topic.Topic, hmacKeys ...interfaces.HmacKey) {
	t.Helper()
	input := interfaces.SubscriptionInput{Topic: tp, HmacKeys: hmacKeys}
	err := l.subscriptions.SubscribeWithMetadata(context.Background(), installationID, []interfaces.SubscriptionInput{input})
	require.NoError(t, err)
}

// buildGroupMessageOriginatorEnvelope creates an OriginatorEnvelope containing a GroupMessageInput.
func buildGroupMessageOriginatorEnvelope(
	t *testing.T,
	nodeID uint32,
	sequenceID uint64,
	timestampNs int64,
	groupID []byte,
	data []byte,
	senderHmac []byte,
	shouldPush bool,
) *envelopesProto.OriginatorEnvelope {
	t.Helper()

	groupInput := &mlsV1.GroupMessageInput{
		Version: &mlsV1.GroupMessageInput_V1_{
			V1: &mlsV1.GroupMessageInput_V1{
				Data:       data,
				SenderHmac: senderHmac,
				ShouldPush: shouldPush,
			},
		},
	}

	clientEnv := &envelopesProto.ClientEnvelope{
		Payload: &envelopesProto.ClientEnvelope_GroupMessage{
			GroupMessage: groupInput,
		},
		Aad: &envelopesProto.AuthenticatedData{
			TargetTopic: topic.NewTopic(topic.TopicKindGroupMessagesV1, groupID).Bytes(),
		},
	}

	payerEnv := testEnvelopes.CreatePayerEnvelope(t, nodeID, clientEnv)
	return testEnvelopes.CreateOriginatorEnvelopeWithTimestamp(t, nodeID, sequenceID, time.Unix(0, timestampNs), payerEnv)
}

// buildWelcomeMessageOriginatorEnvelope creates an OriginatorEnvelope containing a WelcomeMessageInput.
func buildWelcomeMessageOriginatorEnvelope(
	t *testing.T,
	nodeID uint32,
	sequenceID uint64,
	timestampNs int64,
	installationKeyID []byte,
	data []byte,
) *envelopesProto.OriginatorEnvelope {
	t.Helper()

	welcomeInput := &mlsV1.WelcomeMessageInput{
		Version: &mlsV1.WelcomeMessageInput_V1_{
			V1: &mlsV1.WelcomeMessageInput_V1{
				InstallationKey: installationKeyID,
				Data:            data,
			},
		},
	}

	clientEnv := &envelopesProto.ClientEnvelope{
		Payload: &envelopesProto.ClientEnvelope_WelcomeMessage{
			WelcomeMessage: welcomeInput,
		},
		Aad: &envelopesProto.AuthenticatedData{
			TargetTopic: topic.NewTopic(topic.TopicKindWelcomeMessagesV1, installationKeyID).Bytes(),
		},
	}

	payerEnv := testEnvelopes.CreatePayerEnvelope(t, nodeID, clientEnv)
	return testEnvelopes.CreateOriginatorEnvelopeWithTimestamp(t, nodeID, sequenceID, time.Unix(0, timestampNs), payerEnv)
}

// buildPayerReportOriginatorEnvelope creates an OriginatorEnvelope containing a PayerReport (non-convertible).
func buildPayerReportOriginatorEnvelope(
	t *testing.T,
	nodeID uint32,
	sequenceID uint64,
	timestampNs int64,
) *envelopesProto.OriginatorEnvelope {
	t.Helper()
	// Use a PayerReport payload which is non-group/welcome
	reportID := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	clientEnv := testEnvelopes.CreatePayerReportClientEnvelope(nodeID)
	_ = reportID // topic bytes are set by CreatePayerReportClientEnvelope internally
	payerEnv := testEnvelopes.CreatePayerEnvelope(t, nodeID, clientEnv)
	return testEnvelopes.CreateOriginatorEnvelopeWithTimestamp(t, nodeID, sequenceID, time.Unix(0, timestampNs), payerEnv)
}

// TestV4Listener_ProcessGroupMessage_V3Format tests that a GroupMessageInput envelope is converted
// to V3 format and delivered to a V3 installation.
func TestV4Listener_ProcessGroupMessage_V3Format(t *testing.T) {
	mockDelivery := mocks.NewDelivery(t)
	mockDelivery.On("CanDeliver", mock.Anything).Return(true)
	mockDelivery.On("Send", mock.Anything, mock.Anything).Return(nil)

	l := buildV4TestListener(t, mockDelivery)

	groupID := []byte{0x01, 0x02, 0x03, 0x04}
	groupTopic := topic.NewTopic(topic.TopicKindGroupMessagesV1, groupID)

	registerV4Installation(t, l, "inst-v3", interfaces.PayloadFormatV3)
	subscribeV4ToTopic(t, l, "inst-v3", groupTopic)

	env := buildGroupMessageOriginatorEnvelope(t, 1, 1, int64(time.Second), groupID, []byte("data"), nil, true)
	err := l.processOriginatorEnvelope(env)
	require.NoError(t, err)

	mockDelivery.AssertNumberOfCalls(t, "Send", 1)

	// Verify payload format is V3
	sendReqs := getSendRequests(mockDelivery)
	require.Len(t, sendReqs, 1)
	capturedReq := sendReqs[0]
	require.Equal(t, interfaces.PayloadFormatV3, capturedReq.PayloadFormat)
	require.NotEmpty(t, capturedReq.EncryptedMessage)

	// Deserialize and verify it's a V3 GroupMessage
	var groupMsg mlsV1.GroupMessage
	require.NoError(t, proto.Unmarshal(capturedReq.EncryptedMessage, &groupMsg))
	require.NotNil(t, groupMsg.GetV1())
	require.Equal(t, groupID, groupMsg.GetV1().GetGroupId())
}

// TestV4Listener_ProcessGroupMessage_V4Format tests that a GroupMessageInput envelope is delivered
// as raw OriginatorEnvelope bytes to a V4 installation.
func TestV4Listener_ProcessGroupMessage_V4Format(t *testing.T) {
	mockDelivery := mocks.NewDelivery(t)
	mockDelivery.On("CanDeliver", mock.Anything).Return(true)
	mockDelivery.On("Send", mock.Anything, mock.Anything).Return(nil)

	l := buildV4TestListener(t, mockDelivery)

	groupID := []byte{0x05, 0x06, 0x07, 0x08}
	groupTopic := topic.NewTopic(topic.TopicKindGroupMessagesV1, groupID)

	registerV4Installation(t, l, "inst-v4", interfaces.PayloadFormatV4)
	subscribeV4ToTopic(t, l, "inst-v4", groupTopic)

	env := buildGroupMessageOriginatorEnvelope(t, 1, 2, int64(time.Second), groupID, []byte("v4-data"), nil, true)
	err := l.processOriginatorEnvelope(env)
	require.NoError(t, err)

	mockDelivery.AssertNumberOfCalls(t, "Send", 1)

	sendReqs := getSendRequests(mockDelivery)
	require.Len(t, sendReqs, 1)
	capturedReq := sendReqs[0]
	require.Equal(t, interfaces.PayloadFormatV4, capturedReq.PayloadFormat)
	require.NotEmpty(t, capturedReq.EncryptedMessage)

	// Verify it's a valid OriginatorEnvelope
	var origEnv envelopesProto.OriginatorEnvelope
	require.NoError(t, proto.Unmarshal(capturedReq.EncryptedMessage, &origEnv))
}

// TestV4Listener_ProcessWelcomeMessage_V3Format tests that a WelcomeMessageInput envelope is
// converted to V3 format and delivered to a V3 installation.
func TestV4Listener_ProcessWelcomeMessage_V3Format(t *testing.T) {
	mockDelivery := mocks.NewDelivery(t)
	mockDelivery.On("CanDeliver", mock.Anything).Return(true)
	mockDelivery.On("Send", mock.Anything, mock.Anything).Return(nil)

	l := buildV4TestListener(t, mockDelivery)

	installationKeyID := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	welcomeTopic := topic.NewTopic(topic.TopicKindWelcomeMessagesV1, installationKeyID)

	registerV4Installation(t, l, "inst-welcome-v3", interfaces.PayloadFormatV3)
	subscribeV4ToTopic(t, l, "inst-welcome-v3", welcomeTopic)

	env := buildWelcomeMessageOriginatorEnvelope(t, 1, 3, int64(time.Second), installationKeyID, []byte("welcome-data"))
	err := l.processOriginatorEnvelope(env)
	require.NoError(t, err)

	mockDelivery.AssertNumberOfCalls(t, "Send", 1)

	sendReqs := getSendRequests(mockDelivery)
	require.Len(t, sendReqs, 1)
	capturedReq := sendReqs[0]
	require.Equal(t, interfaces.PayloadFormatV3, capturedReq.PayloadFormat)
	require.NotEmpty(t, capturedReq.EncryptedMessage)

	// Deserialize and verify it's a V3 WelcomeMessage
	var welcomeMsg mlsV1.WelcomeMessage
	require.NoError(t, proto.Unmarshal(capturedReq.EncryptedMessage, &welcomeMsg))
	require.NotNil(t, welcomeMsg.GetV1())
	require.Equal(t, installationKeyID, welcomeMsg.GetV1().GetInstallationKey())
}

// TestV4Listener_ProcessGroupMessage_MixedFormats tests that a GroupMessageInput envelope is
// delivered to both V3 and V4 installations, each with the appropriate format.
func TestV4Listener_ProcessGroupMessage_MixedFormats(t *testing.T) {
	mockDelivery := mocks.NewDelivery(t)
	mockDelivery.On("CanDeliver", mock.Anything).Return(true)
	mockDelivery.On("Send", mock.Anything, mock.Anything).Return(nil)

	l := buildV4TestListener(t, mockDelivery)

	groupID := []byte{0x11, 0x22, 0x33, 0x44}
	groupTopic := topic.NewTopic(topic.TopicKindGroupMessagesV1, groupID)

	registerV4Installation(t, l, "inst-mixed-v3", interfaces.PayloadFormatV3)
	registerV4Installation(t, l, "inst-mixed-v4", interfaces.PayloadFormatV4)
	subscribeV4ToTopic(t, l, "inst-mixed-v3", groupTopic)
	subscribeV4ToTopic(t, l, "inst-mixed-v4", groupTopic)

	env := buildGroupMessageOriginatorEnvelope(t, 1, 4, int64(time.Second), groupID, []byte("mixed-data"), nil, true)
	err := l.processOriginatorEnvelope(env)
	require.NoError(t, err)

	mockDelivery.AssertNumberOfCalls(t, "Send", 2)

	formats := map[interfaces.PayloadFormat]int{}
	for _, call := range mockDelivery.Calls {
		if call.Method == "Send" {
			req := call.Arguments.Get(1).(interfaces.SendRequest)
			formats[req.PayloadFormat]++
		}
	}
	require.Equal(t, 1, formats[interfaces.PayloadFormatV3])
	require.Equal(t, 1, formats[interfaces.PayloadFormatV4])
}

// TestV4Listener_SkipNonConvertiblePayload_V3Format tests that a non-group/welcome payload
// is skipped for V3 installations (no delivery).
func TestV4Listener_SkipNonConvertiblePayload_V3Format(t *testing.T) {
	mockDelivery := mocks.NewDelivery(t)
	// No Send calls expected

	l := buildV4TestListener(t, mockDelivery)

	// We need a subscription for the payer report topic. PayerReport uses TopicKindPayerReportsV1.
	payerReportTopic := topic.NewTopic(topic.TopicKindPayerReportsV1, []byte{0x00, 0x00, 0x00, 0x01})

	registerV4Installation(t, l, "inst-payer-v3", interfaces.PayloadFormatV3)
	subscribeV4ToTopic(t, l, "inst-payer-v3", payerReportTopic)

	env := buildPayerReportOriginatorEnvelope(t, 1, 5, int64(time.Second))
	err := l.processOriginatorEnvelope(env)
	require.NoError(t, err)

	mockDelivery.AssertNotCalled(t, "Send")
}

// TestV4Listener_DeliverNonConvertiblePayload_V4Format tests that a non-group/welcome payload
// is still delivered to V4 installations as raw bytes.
func TestV4Listener_DeliverNonConvertiblePayload_V4Format(t *testing.T) {
	mockDelivery := mocks.NewDelivery(t)
	mockDelivery.On("CanDeliver", mock.Anything).Return(true)
	mockDelivery.On("Send", mock.Anything, mock.Anything).Return(nil)

	l := buildV4TestListener(t, mockDelivery)

	payerReportTopic := topic.NewTopic(topic.TopicKindPayerReportsV1, []byte{0x00, 0x00, 0x00, 0x02})

	registerV4Installation(t, l, "inst-payer-v4", interfaces.PayloadFormatV4)
	subscribeV4ToTopic(t, l, "inst-payer-v4", payerReportTopic)

	// Build a payer report envelope targeting this specific topic
	clientEnv := &envelopesProto.ClientEnvelope{
		Payload: &envelopesProto.ClientEnvelope_PayerReport{
			PayerReport: &envelopesProto.PayerReport{},
		},
		Aad: &envelopesProto.AuthenticatedData{
			TargetTopic: payerReportTopic.Bytes(),
		},
	}
	payerEnv := testEnvelopes.CreatePayerEnvelope(t, 1, clientEnv)
	env := testEnvelopes.CreateOriginatorEnvelopeWithTimestamp(t, 1, 6, time.Unix(0, int64(time.Second)), payerEnv)

	err := l.processOriginatorEnvelope(env)
	require.NoError(t, err)

	mockDelivery.AssertNumberOfCalls(t, "Send", 1)

	sendReqs := getSendRequests(mockDelivery)
	require.Len(t, sendReqs, 1)
	require.Equal(t, interfaces.PayloadFormatV4, sendReqs[0].PayloadFormat)
}

// TestV4Listener_ShouldPushFalse_SkipsDelivery tests that a GroupMessageInput with
// ShouldPush=false results in no delivery.
func TestV4Listener_ShouldPushFalse_SkipsDelivery(t *testing.T) {
	mockDelivery := mocks.NewDelivery(t)
	// No Send calls expected

	l := buildV4TestListener(t, mockDelivery)

	groupID := []byte{0x55, 0x66, 0x77, 0x88}
	groupTopic := topic.NewTopic(topic.TopicKindGroupMessagesV1, groupID)

	registerV4Installation(t, l, "inst-nopush-v3", interfaces.PayloadFormatV3)
	subscribeV4ToTopic(t, l, "inst-nopush-v3", groupTopic)

	// ShouldPush = false
	env := buildGroupMessageOriginatorEnvelope(t, 1, 7, int64(time.Second), groupID, []byte("data"), nil, false)
	err := l.processOriginatorEnvelope(env)
	require.NoError(t, err)

	mockDelivery.AssertNotCalled(t, "Send")
}

// TestV4Listener_HmacSenderFiltering tests that when the sender HMAC matches the subscription key,
// delivery is skipped.
func TestV4Listener_HmacSenderFiltering(t *testing.T) {
	mockDelivery := mocks.NewDelivery(t)
	// No Send calls expected for the sender

	l := buildV4TestListener(t, mockDelivery)

	groupID := []byte{0x99, 0xAA, 0xBB, 0xCC}
	groupTopic := topic.NewTopic(topic.TopicKindGroupMessagesV1, groupID)

	// Build HMAC key and compute sender HMAC over the message data
	hmacKey := []byte("test-hmac-key")
	messageData := []byte("group-message-data")
	h := hmac.New(sha256.New, hmacKey)
	h.Write(messageData)
	senderHmac := h.Sum(nil)

	// Compute thirty day period
	timestampNs := int64(time.Second)
	thirtyDayPeriod := int(timestampNs / 1_000_000_000 / 60 / 60 / 24 / 30)

	registerV4Installation(t, l, "inst-hmac-v3", interfaces.PayloadFormatV3)
	subscribeV4ToTopic(t, l, "inst-hmac-v3", groupTopic, interfaces.HmacKey{
		ThirtyDayPeriodsSinceEpoch: thirtyDayPeriod,
		Key:                        hmacKey,
	})

	env := buildGroupMessageOriginatorEnvelope(t, 1, 8, timestampNs, groupID, messageData, senderHmac, true)
	err := l.processOriginatorEnvelope(env)
	require.NoError(t, err)

	// The sender matches — should be filtered
	mockDelivery.AssertNotCalled(t, "Send")
}
