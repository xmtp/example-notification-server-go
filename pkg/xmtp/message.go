package xmtp

import (
	"errors"

	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	messageApi "github.com/xmtp/example-notification-server-go/pkg/proto/message_api/v1"
	messageContents "github.com/xmtp/example-notification-server-go/pkg/proto/message_contents"
	"github.com/xmtp/example-notification-server-go/pkg/topics"
	"google.golang.org/protobuf/proto"
)

func parseConversationMessage(message []byte) (*messageContents.MessageV2, error) {
	var msg messageContents.Message
	err := proto.Unmarshal(message, &msg)
	if err != nil {
		return nil, err
	}
	v2Message := msg.GetV2()
	if v2Message != nil {
		return v2Message, nil
	}
	return nil, errors.New("Not a V1 message")
}

func getContext(env *messageApi.Envelope) interfaces.MessageContext {
	messageType := topics.GetMessageType(env)
	var shouldPush *bool
	var hmacInputs, senderHmac *[]byte
	if messageType == topics.V2Conversation {
		if parsed, err := parseConversationMessage(env.Message); err == nil {
			shouldPush = parsed.ShouldPush
			hmacInputs = &parsed.HeaderBytes
			if len(parsed.SenderHmac) > 0 {
				senderHmac = &parsed.SenderHmac
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
