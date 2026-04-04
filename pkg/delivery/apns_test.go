package delivery

import (
	"encoding/json"
	"testing"

	"github.com/sideshow/apns2/payload"
	"github.com/stretchr/testify/require"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	v1 "github.com/xmtp/xmtpd/pkg/proto/message_api/v1"
	"github.com/xmtp/example-notification-server-go/pkg/topics"
)

func Test_ApnsDelivery_BuildNotification_TopicField(t *testing.T) {
	parsed, err := topics.ParseV3Topic("/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto")
	require.NoError(t, err)

	a := ApnsDelivery{opts: options.ApnsOptions{Topic: "com.example.app"}}
	req := interfaces.SendRequest{
		Message: &v1.Envelope{Message: []byte("test")},
		Subscription: interfaces.Subscription{
			TopicV4: parsed,
			Topic:   topics.TopicToString(parsed),
		},
		Installation: interfaces.Installation{
			DeliveryMechanism: interfaces.DeliveryMechanism{Token: "device-token"},
		},
		MessageContext: interfaces.MessageContext{MessageType: topics.V3Conversation},
	}

	notification := a.buildNotification(req)
	payloadBytes, err := notification.Payload.(*payload.Payload).MarshalJSON()
	require.NoError(t, err)

	var p map[string]interface{}
	require.NoError(t, json.Unmarshal(payloadBytes, &p))
	require.Equal(t, "/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto", p["topic"])
	require.NotContains(t, p, "topicBytesB64")
	require.Equal(t, "device-token", notification.DeviceToken)
	require.Equal(t, "com.example.app", notification.Topic)
}
