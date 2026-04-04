package topics

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xmtp/xmtpd/pkg/topic"
)

func Test_ParseV3Topic_GroupMessage(t *testing.T) {
	result, err := ParseV3Topic("/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto")
	require.NoError(t, err)
	require.Equal(t, topic.TopicKindGroupMessagesV1, result.Kind())
	require.Equal(t, []byte{0x24, 0xce, 0x39, 0xd6, 0x60, 0x60, 0x0b, 0x3a, 0x98, 0xad, 0xff, 0x30, 0x75, 0xb6, 0xd1, 0xf4}, result.Identifier())
}

func Test_ParseV3Topic_WelcomeMessage(t *testing.T) {
	result, err := ParseV3Topic("/xmtp/mls/1/w-f3ac64eba2272334124975d673374bdd64a5535bf5f7b48ac5608ff499444be0/proto")
	require.NoError(t, err)
	require.Equal(t, topic.TopicKindWelcomeMessagesV1, result.Kind())
}

func Test_ParseV3Topic_MixedCase(t *testing.T) {
	result, err := ParseV3Topic("/xmtp/mls/1/g-24CE39D660600B3A98ADFF3075B6D1F4/proto")
	require.NoError(t, err)
	require.Equal(t, topic.TopicKindGroupMessagesV1, result.Kind())
}

func Test_ParseV3Topic_InvalidPattern(t *testing.T) {
	_, err := ParseV3Topic("/xmtp/mls/1/x-abc123/proto")
	require.Error(t, err)
}

func Test_ParseV3Topic_NotHex(t *testing.T) {
	_, err := ParseV3Topic("/xmtp/mls/1/g-zzzz/proto")
	require.Error(t, err)
}

func Test_ParseV3Topic_MissingSuffix(t *testing.T) {
	_, err := ParseV3Topic("/xmtp/mls/1/g-24ce39d660600b3a")
	require.Error(t, err)
}

func Test_TopicToString_GroupMessage(t *testing.T) {
	identifier, _ := hex.DecodeString("24ce39d660600b3a98adff3075b6d1f4")
	tp := topic.NewTopic(topic.TopicKindGroupMessagesV1, identifier)
	result := TopicToString(tp)
	require.Equal(t, "/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto", result)
}

func Test_TopicToString_WelcomeMessage(t *testing.T) {
	identifier, _ := hex.DecodeString("f3ac64eba2272334")
	tp := topic.NewTopic(topic.TopicKindWelcomeMessagesV1, identifier)
	result := TopicToString(tp)
	require.Equal(t, "/xmtp/mls/1/w-f3ac64eba2272334/proto", result)
}

func Test_TopicToString_UnknownKind(t *testing.T) {
	identifier, _ := hex.DecodeString("abcd")
	tp := topic.NewTopic(topic.TopicKindIdentityUpdatesV1, identifier)
	result := TopicToString(tp)
	require.Equal(t, "", result)
}

func Test_TopicToString_Roundtrip(t *testing.T) {
	original := "/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto"
	parsed, err := ParseV3Topic(original)
	require.NoError(t, err)
	require.Equal(t, original, TopicToString(parsed))
}

func Test_GetMessageTypeFromTopic_Group(t *testing.T) {
	tp := topic.NewTopic(topic.TopicKindGroupMessagesV1, []byte{0x01})
	require.Equal(t, V3Conversation, GetMessageTypeFromTopic(tp))
}

func Test_GetMessageTypeFromTopic_Welcome(t *testing.T) {
	tp := topic.NewTopic(topic.TopicKindWelcomeMessagesV1, []byte{0x01})
	require.Equal(t, V3Welcome, GetMessageTypeFromTopic(tp))
}

func Test_GetMessageTypeFromTopic_Unknown(t *testing.T) {
	tp := topic.NewTopic(topic.TopicKindIdentityUpdatesV1, []byte{0x01})
	require.Equal(t, Unknown, GetMessageTypeFromTopic(tp))
}

func Test_GetMessageTypeFromTopic_Nil(t *testing.T) {
	require.Equal(t, Unknown, GetMessageTypeFromTopic(nil))
}
