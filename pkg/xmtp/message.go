package xmtp

import (
	"errors"

	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	messageApi "github.com/xmtp/example-notification-server-go/pkg/proto/message_api/v1"
	mlsV1 "github.com/xmtp/example-notification-server-go/pkg/proto/mls/api/v1"
	"github.com/xmtp/example-notification-server-go/pkg/topics"
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

func getContext(env *messageApi.Envelope) interfaces.MessageContext {
	messageType := topics.GetMessageType(env)
	var shouldPush *bool
	var hmacInputs, senderHmac *[]byte

	if messageType == topics.V3Conversation {
		if message, err := parseGroupMessage(env.Message); err == nil {
			shouldPush = &message.ShouldPush

			hmacInputs = &message.Data
			if len(message.SenderHmac) > 0 {
				senderHmac = &message.SenderHmac
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
