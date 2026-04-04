package topics

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/xmtp/xmtpd/pkg/topic"
)

const V3_PREFIX = "/xmtp/mls/1/"

func ParseV3Topic(topicStr string) (*topic.Topic, error) {
	topicStr = strings.TrimPrefix(topicStr, V3_PREFIX)
	if !strings.HasSuffix(topicStr, "/proto") {
		return nil, fmt.Errorf("invalid V3 topic: missing /proto suffix")
	}
	topicStr = strings.TrimSuffix(topicStr, "/proto")

	prefix, identifier, hasPrefix := strings.Cut(topicStr, "-")
	if !hasPrefix || identifier == "" {
		return nil, fmt.Errorf("invalid V3 topic: missing prefix or identifier")
	}

	var kind topic.TopicKind
	switch prefix {
	case "g":
		kind = topic.TopicKindGroupMessagesV1
	case "w":
		kind = topic.TopicKindWelcomeMessagesV1
	default:
		return nil, fmt.Errorf("invalid V3 topic: unknown prefix %q", prefix)
	}

	identifierBytes, err := hex.DecodeString(strings.ToLower(identifier))
	if err != nil {
		return nil, fmt.Errorf("invalid V3 topic: bad hex identifier: %w", err)
	}

	return topic.NewTopic(kind, identifierBytes), nil
}

func TopicToString(t *topic.Topic) string {
	if t == nil {
		return ""
	}
	var prefix string
	switch t.Kind() {
	case topic.TopicKindGroupMessagesV1:
		prefix = "g"
	case topic.TopicKindWelcomeMessagesV1:
		prefix = "w"
	default:
		return ""
	}
	return V3_PREFIX + prefix + "-" + hex.EncodeToString(t.Identifier()) + "/proto"
}

func GetMessageTypeFromTopic(t *topic.Topic) MessageType {
	if t == nil {
		return Unknown
	}
	switch t.Kind() {
	case topic.TopicKindGroupMessagesV1:
		return V3Conversation
	case topic.TopicKindWelcomeMessagesV1:
		return V3Welcome
	default:
		return Unknown
	}
}
