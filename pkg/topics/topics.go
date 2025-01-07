package topics

import (
	"strings"

	v1 "github.com/xmtp/example-notification-server-go/pkg/proto/message_api/v1"
)

const V1_PREFIX = "/xmtp/0/"
const V3_PREFIX = "/xmtp/mls/1/"

var messageTypeByPrefix = map[string]MessageType{
	"test":         Test,
	"privatestore": Private,
	"contact":      Contact,
	"intro":        V1Intro,
	"dm":           V1Conversation,
	"invite":       V2Invite,
	"m":            V2Conversation,
	"g":            V3Conversation,
	"w":            V3Welcome,
}

func GetMessageType(env *v1.Envelope) MessageType {
	topic := env.ContentTopic
	if strings.HasPrefix(topic, "test-") {
		return Test
	}
	topic = strings.TrimPrefix(topic, V1_PREFIX)
	topic = strings.TrimPrefix(topic, V3_PREFIX)
	prefix, _, hasPrefix := strings.Cut(topic, "-")
	if hasPrefix {
		if category, found := messageTypeByPrefix[prefix]; found {
			return category
		}
	}

	return Unknown
}
