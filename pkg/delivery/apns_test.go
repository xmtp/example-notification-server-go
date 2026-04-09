package delivery

import (
	"encoding/json"
	"testing"

	"github.com/sideshow/apns2/payload"
	"github.com/stretchr/testify/require"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	"github.com/xmtp/example-notification-server-go/pkg/topics"
)

const deliveryTestTopic = "/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto"

func buildDeliveryRequest(t *testing.T, payloadFormat interfaces.PayloadFormat) interfaces.SendRequest {
	t.Helper()

	parsed, err := topics.ParseV3Topic(deliveryTestTopic)
	require.NoError(t, err)
	topicStr := topics.TopicToString(parsed)
	req := interfaces.SendRequest{
		Topic:            topicStr,
		EncryptedMessage: []byte("test"),
		PayloadFormat:    payloadFormat,
		Subscription: interfaces.Subscription{
			TopicV4: parsed,
			Topic:   topicStr,
		},
		Installation: interfaces.Installation{
			DeliveryMechanism: interfaces.DeliveryMechanism{Token: "device-token"},
		},
		MessageContext: interfaces.MessageContext{MessageType: topics.V3Conversation},
	}
	if payloadFormat == interfaces.PayloadFormatV4 {
		req.TopicBytesB64 = topics.TopicToBase64(parsed)
	}
	return req
}

func TestApns_PayloadIncludesPayloadFormat(t *testing.T) {
	a := ApnsDelivery{opts: options.ApnsOptions{Topic: "com.example.app"}}
	req := buildDeliveryRequest(t, interfaces.PayloadFormatV3)

	notification := a.buildNotification(req)
	payloadBytes, err := notification.Payload.(*payload.Payload).MarshalJSON()
	require.NoError(t, err)

	var p map[string]interface{}
	require.NoError(t, json.Unmarshal(payloadBytes, &p))
	require.Equal(t, "v3", p["payloadFormat"])
}

func Test_ApnsDelivery_BuildNotification_TopicField(t *testing.T) {
	a := ApnsDelivery{opts: options.ApnsOptions{Topic: "com.example.app"}}
	req := buildDeliveryRequest(t, interfaces.PayloadFormatV3)

	notification := a.buildNotification(req)
	payloadBytes, err := notification.Payload.(*payload.Payload).MarshalJSON()
	require.NoError(t, err)

	var p map[string]interface{}
	require.NoError(t, json.Unmarshal(payloadBytes, &p))
	require.Equal(t, deliveryTestTopic, p["topic"])
	require.NotContains(t, p, "topicBytesB64")
	require.Equal(t, "device-token", notification.DeviceToken)
	require.Equal(t, "com.example.app", notification.Topic)
}

func Test_ApnsDelivery_BuildNotification_V4TopicBytesB64(t *testing.T) {
	a := ApnsDelivery{opts: options.ApnsOptions{Topic: "com.example.app"}}
	req := buildDeliveryRequest(t, interfaces.PayloadFormatV4)

	notification := a.buildNotification(req)
	payloadBytes, err := notification.Payload.(*payload.Payload).MarshalJSON()
	require.NoError(t, err)

	var p map[string]interface{}
	require.NoError(t, json.Unmarshal(payloadBytes, &p))
	require.Equal(t, deliveryTestTopic, p["topic"])
	require.Equal(t, req.TopicBytesB64, p["topicBytesB64"])
	require.Equal(t, "v4", p["payloadFormat"])
}
