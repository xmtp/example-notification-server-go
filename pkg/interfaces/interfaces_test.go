package interfaces

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	proto "github.com/xmtp/example-notification-server-go/pkg/proto/notifications/v1"
	"github.com/xmtp/xmtpd/pkg/topic"
)

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

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &result))
	require.Equal(t, "/xmtp/mls/1/g-24ce/proto", result["topic"])
	require.NotContains(t, result, "topicBytesB64")
	require.Equal(t, true, result["is_silent"])
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

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &result))

	// Must have nested message object with content_topic and message
	msg, ok := result["message"].(map[string]interface{})
	require.True(t, ok, "expected 'message' object in JSON")
	require.Equal(t, "/xmtp/mls/1/w-test/proto", msg["content_topic"])
	require.NotEmpty(t, msg["message"])

	// Must NOT have top-level topic or encrypted_message
	_, hasTopic := result["topic"]
	require.False(t, hasTopic, "top-level 'topic' should not exist")
	_, hasEncMsg := result["encrypted_message"]
	require.False(t, hasEncMsg, "top-level 'encrypted_message' should not exist")

	// Other fields
	require.Equal(t, "abc123", result["idempotency_key"])
	mc := result["message_context"].(map[string]interface{})
	require.Equal(t, "v3-welcome", mc["message_type"])
}

func Test_Subscription_MarshalJSON_TopicV4NotSerialized(t *testing.T) {
	tp := topic.NewTopic(topic.TopicKindGroupMessagesV1, []byte{0x24, 0xce})
	sub := Subscription{TopicV4: tp, Topic: "", IsSilent: false}
	data, err := json.Marshal(sub)
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &result))
	require.Equal(t, "", result["topic"])
	// TopicV4 is json:"-", should not appear in output
	_, hasTopicV4 := result["topicV4"]
	require.False(t, hasTopicV4)
}
