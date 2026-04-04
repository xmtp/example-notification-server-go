package delivery

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/topics"
)

func TestFcm_DataIncludesPayloadFormat(t *testing.T) {
	parsed, err := topics.ParseV3Topic("/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto")
	require.NoError(t, err)

	topicStr := topics.TopicToString(parsed)
	req := interfaces.SendRequest{
		Topic:            topicStr,
		EncryptedMessage: []byte("test"),
		PayloadFormat:    interfaces.PayloadFormatV3,
		Subscription: interfaces.Subscription{
			TopicV4: parsed,
			Topic:   topicStr,
		},
		MessageContext: interfaces.MessageContext{MessageType: topics.V3Conversation},
	}

	data := buildFcmData(req)
	require.Equal(t, "v3", data["payloadFormat"])
}

func Test_BuildFcmData_TopicField(t *testing.T) {
	parsed, err := topics.ParseV3Topic("/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto")
	require.NoError(t, err)

	topicStr := topics.TopicToString(parsed)
	req := interfaces.SendRequest{
		Topic:            topicStr,
		EncryptedMessage: []byte("test"),
		Subscription: interfaces.Subscription{
			TopicV4: parsed,
			Topic:   topicStr,
		},
		MessageContext: interfaces.MessageContext{MessageType: topics.V3Conversation},
	}

	data := buildFcmData(req)
	require.Equal(t, "/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto", data["topic"])
	require.NotContains(t, data, "topicBytesB64")
	require.Equal(t, base64.StdEncoding.EncodeToString([]byte("test")), data["encryptedMessage"])
	require.Equal(t, "v3-conversation", data["messageType"])
}
