package topics

import (
	"strings"
)

const V3_PREFIX = "/xmtp/mls/1/"

var messageTypeByPrefix = map[string]MessageType{
	"test": Test,
	"g":    V3Conversation,
	"w":    V3Welcome,
}

func GetMessageType(topic string) MessageType {

	if strings.HasPrefix(topic, "test-") {
		return Test
	}

	topic = strings.TrimPrefix(topic, V3_PREFIX)
	prefix, _, hasPrefix := strings.Cut(topic, "-")
	if hasPrefix {
		if category, found := messageTypeByPrefix[prefix]; found {
			return category
		}
	}

	return Unknown
}
