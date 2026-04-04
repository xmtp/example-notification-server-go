package interfaces

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xmtp/xmtpd/pkg/topic"
)

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
