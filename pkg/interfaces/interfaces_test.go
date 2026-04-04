package interfaces

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xmtp/xmtpd/pkg/topic"
)

func Test_Subscription_MarshalJSON_BothFields(t *testing.T) {
	tp := topic.NewTopic(topic.TopicKindGroupMessagesV1, []byte{0x24, 0xce})
	sub := Subscription{
		Topic:       tp,
		TopicString: "/xmtp/mls/1/g-24ce/proto",
		IsSilent:    true,
	}
	data, err := json.Marshal(sub)
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &result))
	require.Equal(t, "/xmtp/mls/1/g-24ce/proto", result["topic"])
	require.Equal(t, base64.StdEncoding.EncodeToString(tp.Bytes()), result["topicBytesB64"])
	require.Equal(t, true, result["is_silent"])
}

func Test_Subscription_MarshalJSON_NilTopic(t *testing.T) {
	sub := Subscription{TopicString: "", IsSilent: false}
	data, err := json.Marshal(sub)
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &result))
	require.Equal(t, "", result["topic"])
	require.Equal(t, "", result["topicBytesB64"])
}
