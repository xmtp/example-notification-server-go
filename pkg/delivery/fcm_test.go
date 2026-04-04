package delivery

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	v1 "github.com/xmtp/example-notification-server-go/pkg/proto/message_api/v1"
	"github.com/xmtp/example-notification-server-go/pkg/topics"
)

func Test_BuildFcmData_TopicField(t *testing.T) {
	parsed, err := topics.ParseV3Topic("/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto")
	require.NoError(t, err)

	req := interfaces.SendRequest{
		Message: &v1.Envelope{Message: []byte("test")},
		Subscription: interfaces.Subscription{
			TopicV4: parsed,
			Topic:   topics.TopicToString(parsed),
		},
		MessageContext: interfaces.MessageContext{MessageType: topics.V3Conversation},
	}

	data := buildFcmData(req)
	require.Equal(t, "/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto", data["topic"])
	require.NotContains(t, data, "topicBytesB64")
	require.Equal(t, base64.StdEncoding.EncodeToString([]byte("test")), data["encryptedMessage"])
	require.Equal(t, "v3-conversation", data["messageType"])
}
