package topics

import (
	"strings"

	v1 "github.com/xmtp/example-notification-server-go/pkg/proto/message_api/v1"
	"go.uber.org/zap"
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

func GetMessageType(env *v1.Envelope, logger *zap.Logger) MessageType {
	topic := env.ContentTopic
	logger.Info("Determining message type", zap.String("contentTopic", topic))

	if strings.HasPrefix(topic, "test-") {
		logger.Info("Detected test topic", zap.String("contentTopic", topic))
		return Test
	}

	originalTopic := topic
	topic = strings.TrimPrefix(topic, V1_PREFIX)
	topic = strings.TrimPrefix(topic, V3_PREFIX)
	logger.Info("Trimmed topic prefixes", zap.String("originalTopic", originalTopic), zap.String("trimmedTopic", topic))

	prefix, _, hasPrefix := strings.Cut(topic, "-")
	if hasPrefix {
		logger.Info("Extracted prefix", zap.String("prefix", prefix))
		if category, found := messageTypeByPrefix[prefix]; found {
			logger.Info("Matched prefix to category", zap.String("prefix", prefix), zap.String("category", string(category)))
			return category
		}
		logger.Info("Prefix not found in messageTypeByPrefix", zap.String("prefix", prefix))
	}

	logger.Info("Message type is unknown", zap.String("contentTopic", topic))
	return Unknown
}
