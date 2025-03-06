package xmtp

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	messageApi "github.com/xmtp/example-notification-server-go/pkg/proto/message_api/v1"
	"github.com/xmtp/example-notification-server-go/pkg/topics"
	"google.golang.org/protobuf/proto"
)

type rawFixture struct {
	Kind     string
	Envelope string
	HmacKey  string
}

func getRawFixture(t *testing.T, kind string) *rawFixture {
	data, err := os.ReadFile("../../fixtures/envelopes.json")
	require.NoError(t, err)
	fixtures := []rawFixture{}
	err = json.Unmarshal(data, &fixtures)
	require.NoError(t, err)

	for _, fixture := range fixtures {
		if fixture.Kind == kind {
			return &fixture
		}
	}

	t.Fail()
	return nil
}

func getEnvelope(t *testing.T, fixture *rawFixture) *messageApi.Envelope {
	envelopeBytes, err := hex.DecodeString(fixture.Envelope)
	require.NoError(t, err)

	envelope := &messageApi.Envelope{}
	err = proto.Unmarshal(envelopeBytes, envelope)
	require.NoError(t, err)

	return envelope
}

func getHmacKey(t *testing.T, fixture *rawFixture) []byte {
	hmacKey, err := hex.DecodeString(fixture.HmacKey)
	require.NoError(t, err)

	return hmacKey
}

func Test_IdentifyV3Conversation(t *testing.T) {
	rawFixture := getRawFixture(t, "v3-conversation")
	envelope := getEnvelope(t, rawFixture)
	hmacKey := getHmacKey(t, rawFixture)
	context := getContext(envelope)
	require.False(t, context.IsSender(hmacKey))
	require.True(t, *context.ShouldPush)
	require.Equal(t, context.MessageType, topics.V3Conversation)

	wrongKey := []byte("foo")
	contextWithWrongKey := getContext(envelope)
	require.False(t, contextWithWrongKey.IsSender(wrongKey))
	require.True(t, *contextWithWrongKey.ShouldPush)
	require.Equal(t, contextWithWrongKey.MessageType, topics.V3Conversation)
}
