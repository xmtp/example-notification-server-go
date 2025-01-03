package xmtp

import (
	"errors"

	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	messageApi "github.com/xmtp/example-notification-server-go/pkg/proto/message_api/v1"
	messageContents "github.com/xmtp/example-notification-server-go/pkg/proto/message_contents"
	mlsV1 "github.com/xmtp/example-notification-server-go/pkg/proto/mls/api/v1"
	"github.com/xmtp/example-notification-server-go/pkg/topics"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func parseGroupMessage(groupMessage []byte, logger *zap.Logger) (*mlsV1.GroupMessage_V1, error) {
	logger.Info("Parsing group message", zap.Int("messageLength", len(groupMessage)))
	
	var msg mlsV1.GroupMessage
	err := proto.Unmarshal(groupMessage, &msg)
	if err != nil {
		logger.Error("Failed to unmarshal message", zap.Error(err))
		return nil, err
	}

	v1Message := msg.GetV1()

	if v1Message == nil {
		logger.Info("Parsed V2 message successfully")
		return nil, errors.New("Not a V1 message")
	}

	return v1Message, nil
}

func parseConversationMessage(message []byte, logger *zap.Logger) (*messageContents.MessageV2, error) {
	logger.Info("Parsing conversation message", zap.Int("messageLength", len(message)))

	var msg messageContents.Message
	err := proto.Unmarshal(message, &msg)
	if err != nil {
		logger.Error("Failed to unmarshal message", zap.Error(err))
		return nil, err
	}

	v2Message := msg.GetV2()
	if v2Message != nil {
		logger.Info("Parsed V2 message successfully")
		return v2Message, nil
	}

	logger.Warn("Not a V1 message")
	return nil, errors.New("Not a V1 message")
}

func getContext(env *messageApi.Envelope, logger *zap.Logger) interfaces.MessageContext {
	logger.Info("Getting topic from envelope", zap.String("envelopeID", env.ContentTopic))
	logger.Info("Getting message from envelope", zap.String("envelopeID", string(env.Message)))

	messageType := topics.GetMessageType(env, logger)
	logger.Info("Determined message type", zap.String("messageType", string(messageType)))

	var shouldPush *bool
	var hmacInputs, senderHmac *[]byte

	if messageType == topics.V2Conversation {
		logger.Info("Processing V2 conversation message")
		if parsed, err := parseConversationMessage(env.Message, logger); err == nil {
			shouldPush = parsed.ShouldPush
			hmacInputs = &parsed.HeaderBytes
			if len(parsed.SenderHmac) > 0 {
				senderHmac = &parsed.SenderHmac
				logger.Info("Set sender HMAC for V2 conversation")
			}
		} else {
			logger.Error("Failed to parse V2 conversation message", zap.Error(err))
		}
	} else if messageType == topics.V3Conversation {
		logger.Info("Processing V3 conversation message")
		if message, err := parseGroupMessage(env.Message, logger); err == nil {
			shouldPush = new(bool)
			*shouldPush = true
			hmacInputs = &message.Data
			logger.Info("Processing V3 conversation message", zap.ByteString("hmacInputs", *hmacInputs))

			if len(message.SenderHmac) > 0 {
				senderHmac = &message.SenderHmac
				logger.Info("Set sender HMAC for V3 conversation", zap.ByteString("senderHmac", *senderHmac))
			} else {
				logger.Warn("Sender HMAC is missing or empty", zap.ByteString("hmacInputs", *hmacInputs))
				if message.SenderHmac == nil {
					logger.Warn("SenderHmac field is nil")
				} else {
					logger.Warn("SenderHmac field is present but empty", zap.Int("length", len(message.SenderHmac)))
				}
			}
		} else {
			logger.Error("Failed to parse V3 conversation message", zap.Error(err))
		}
	}

	context := interfaces.MessageContext{
		MessageType: messageType,
		ShouldPush:  shouldPush,
		HmacInputs:  hmacInputs,
		SenderHmac:  senderHmac,
	}
	logger.Info("Constructed MessageContext", zap.Any("context", context))
	return context
}
