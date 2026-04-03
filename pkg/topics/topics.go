package topics

import (
	"errors"
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

func GetTopicID(topic string) (string, error) {

	// We have a plain topic ID - just keep it as is.
	if !strings.HasPrefix(topic, V3CommonPrefix) {
		return topic, nil
	}

	// Old group message.
	groupTopic, ok := strings.CutPrefix(topic, V3GroupPrefix)
	if ok {
		return groupTopic, nil
	}

	// Old welcome message.
	welcomeTopic, ok := strings.CutPrefix(topic, V3WelcomeMessagePrefix)
	if ok {
		return welcomeTopic, nil
	}

	return "", errors.New("unsupported topic")
}
