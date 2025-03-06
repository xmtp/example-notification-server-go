package xmtp

import (
	"encoding/hex"
	"errors"

	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	messageApi "github.com/xmtp/example-notification-server-go/pkg/proto/message_api/v1"
	mlsV1 "github.com/xmtp/example-notification-server-go/pkg/proto/mls/api/v1"
	"github.com/xmtp/example-notification-server-go/pkg/topics"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func parseGroupMessage(groupMessage []byte) (*mlsV1.GroupMessage_V1, error) {
	var msg mlsV1.GroupMessage
	err := proto.Unmarshal(groupMessage, &msg)
	if err != nil {
		return nil, err
	}

	v1Message := msg.GetV1()

	if v1Message == nil {
		return nil, errors.New("Not a V1 message")
	}

	return v1Message, nil
}

func getContext(env *messageApi.Envelope, logger *zap.Logger) interfaces.MessageContext {
	messageType := topics.GetMessageType(env)
	var shouldPush *bool
	var hmacInputs, senderHmac *[]byte

	if messageType == topics.V3Conversation {
		if message, err := parseGroupMessage(env.Message); err == nil {
			logger.Info("MESSAGE PARSED LOPI",
				zap.Any("should_push", message.ShouldPush),
				zap.Any("hmac", message.SenderHmac),

			)
			shouldPush = &message.ShouldPush

			hmacInputs = &message.Data
			if len(message.SenderHmac) > 0 {
				senderHmac = &message.SenderHmac
			}

			// Convert `shouldPush` to a readable value for logging
			shouldPushVal := false
			if shouldPush != nil {
				shouldPushVal = *shouldPush
			}

			logger.Info("Processing v3-conversation message",
				zap.Bool("shouldPush", shouldPushVal),
			)

			// Marshal the envelope for logging
			envelopeBytes, err := proto.Marshal(env)
			if err != nil {
				logger.Error("Failed to marshal envelope", zap.Error(err))
			} else {
				logger.Info("Envelope Data",
					zap.String("kind", "v3-conversation"),
					zap.String("envelope", hex.EncodeToString(envelopeBytes)),
					zap.Bool("should_push", shouldPushVal),
				)
			}
		}
	}

	return interfaces.MessageContext{
		MessageType: messageType,
		ShouldPush:  shouldPush,
		HmacInputs:  hmacInputs,
		SenderHmac:  senderHmac,
	}
}
