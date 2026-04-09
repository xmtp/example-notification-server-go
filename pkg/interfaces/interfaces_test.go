package interfaces

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	proto "github.com/xmtp/example-notification-server-go/pkg/proto/notifications/v1"
	"github.com/xmtp/xmtpd/pkg/topic"
)

type subscriptionJSON struct {
	Topic         string `json:"topic"`
	IsSilent      bool   `json:"is_silent"`
	TopicBytesB64 string `json:"topicBytesB64"`
}

type sendRequestJSONFixture struct {
	IdempotencyKey string `json:"idempotency_key"`
	Message        struct {
		ContentTopic string `json:"content_topic"`
		Message      string `json:"message"`
	} `json:"message"`
	MessageContext struct {
		MessageType string `json:"message_type"`
	} `json:"message_context"`
	Topic            string `json:"topic"`
	EncryptedMessage string `json:"encrypted_message"`
	TopicBytesB64    string `json:"topicBytesB64"`
}

func TestPayloadFormat_String(t *testing.T) {
	require.Equal(t, "v3", PayloadFormatV3.String())
	require.Equal(t, "v4", PayloadFormatV4.String())
	require.Equal(t, "unspecified", PayloadFormatUnspecified.String())
}

func TestPayloadFormat_FromProto(t *testing.T) {
	require.Equal(t, PayloadFormatV3, PayloadFormatFromProto(proto.PayloadFormat_PAYLOAD_FORMAT_V3))
	require.Equal(t, PayloadFormatV4, PayloadFormatFromProto(proto.PayloadFormat_PAYLOAD_FORMAT_V4))
	require.Equal(t, PayloadFormatUnspecified, PayloadFormatFromProto(proto.PayloadFormat_PAYLOAD_FORMAT_UNSPECIFIED))
}

func TestPayloadFormat_ToProto(t *testing.T) {
	require.Equal(t, proto.PayloadFormat_PAYLOAD_FORMAT_V3, PayloadFormatV3.ToProto())
	require.Equal(t, proto.PayloadFormat_PAYLOAD_FORMAT_V4, PayloadFormatV4.ToProto())
}

func TestNormalizePayloadFormat(t *testing.T) {
	require.Equal(t, PayloadFormatV3, NormalizePayloadFormat(PayloadFormatUnspecified))
	require.Equal(t, PayloadFormatV3, NormalizePayloadFormat(PayloadFormatV3))
	require.Equal(t, PayloadFormatV4, NormalizePayloadFormat(PayloadFormatV4))
}

func TestPayloadFormat_ValidateForListener(t *testing.T) {
	require.NoError(t, PayloadFormatV3.ValidateForListener(ListenerTypeV3))
	require.NoError(t, PayloadFormatV3.ValidateForListener(ListenerTypeV4))
	require.NoError(t, PayloadFormatV4.ValidateForListener(ListenerTypeV4))
	require.Error(t, PayloadFormatV4.ValidateForListener(ListenerTypeV3))
	require.NoError(t, PayloadFormatUnspecified.ValidateForListener(ListenerTypeV3)) // defaults to V3
}

func Test_Subscription_MarshalJSON_TopicOnly(t *testing.T) {
	tp := topic.NewTopic(topic.TopicKindGroupMessagesV1, []byte{0x24, 0xce})
	sub := Subscription{
		TopicV4:  tp,
		Topic:    "/xmtp/mls/1/g-24ce/proto",
		IsSilent: true,
	}
	data, err := json.Marshal(sub)
	require.NoError(t, err)

	var result subscriptionJSON
	require.NoError(t, json.Unmarshal(data, &result))
	require.Equal(t, "/xmtp/mls/1/g-24ce/proto", result.Topic)
	require.Empty(t, result.TopicBytesB64)
	require.True(t, result.IsSilent)
}

func TestSendRequest_MarshalJSON_BackwardCompatible(t *testing.T) {
	req := SendRequest{
		IdempotencyKey:   "abc123",
		Topic:            "/xmtp/mls/1/w-test/proto",
		EncryptedMessage: []byte("encrypted-data"),
		PayloadFormat:    PayloadFormatV3,
		MessageContext:   MessageContext{MessageType: "v3-welcome"},
		Installation: Installation{
			Id:                "install-1",
			DeliveryMechanism: DeliveryMechanism{Kind: "apns", Token: "token"},
		},
		Subscription: Subscription{
			Topic:    "/xmtp/mls/1/w-test/proto",
			IsSilent: true,
		},
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var result sendRequestJSONFixture
	require.NoError(t, json.Unmarshal(data, &result))

	require.Equal(t, "/xmtp/mls/1/w-test/proto", result.Message.ContentTopic)
	require.NotEmpty(t, result.Message.Message)

	require.Empty(t, result.Topic)
	require.Empty(t, result.EncryptedMessage)
	require.Equal(t, "abc123", result.IdempotencyKey)
	require.Equal(t, "v3-welcome", result.MessageContext.MessageType)
}

func TestSendRequest_MarshalJSON_TopicBytesIncluded(t *testing.T) {
	req := SendRequest{
		IdempotencyKey:   "abc123",
		Topic:            "/xmtp/mls/1/g-01020304/proto",
		TopicBytesB64:       "AQECBA==",
		EncryptedMessage: []byte("encrypted-data"),
		PayloadFormat:    PayloadFormatV4,
		MessageContext:   MessageContext{MessageType: "v3-conversation"},
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var result sendRequestJSONFixture
	require.NoError(t, json.Unmarshal(data, &result))

	require.Equal(t, "/xmtp/mls/1/g-01020304/proto", result.Message.ContentTopic)
	require.Equal(t, "AQECBA==", result.TopicBytesB64)
}

func TestSendRequest_MarshalJSON_TopicBytesOmittedWhenEmpty(t *testing.T) {
	req := SendRequest{
		IdempotencyKey:   "abc123",
		Topic:            "/xmtp/mls/1/g-01020304/proto",
		EncryptedMessage: []byte("encrypted-data"),
		PayloadFormat:    PayloadFormatV3,
		MessageContext:   MessageContext{MessageType: "v3-conversation"},
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var result sendRequestJSONFixture
	require.NoError(t, json.Unmarshal(data, &result))

	require.Empty(t, result.TopicBytesB64)
	// Also verify it's not in the raw JSON at all
	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &raw))
	require.NotContains(t, raw, "topicBytesB64")
}

func Test_Subscription_MarshalJSON_TopicV4NotSerialized(t *testing.T) {
	tp := topic.NewTopic(topic.TopicKindGroupMessagesV1, []byte{0x24, 0xce})
	sub := Subscription{TopicV4: tp, Topic: "", IsSilent: false}
	data, err := json.Marshal(sub)
	require.NoError(t, err)

	var result subscriptionJSON
	require.NoError(t, json.Unmarshal(data, &result))
	require.Empty(t, result.Topic)
	require.Empty(t, result.TopicBytesB64)
}
