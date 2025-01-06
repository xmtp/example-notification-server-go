package xmtp

import (
	"errors"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	messageApi "github.com/xmtp/example-notification-server-go/pkg/proto/message_api/v1"
	messageContents "github.com/xmtp/example-notification-server-go/pkg/proto/message_contents"
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

func parseConversationMessage(message []byte) (*messageContents.MessageV2, error) {
	fmt.Println("parseConversationMessage")
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
	fmt.Println("1")
	messageType := topics.GetMessageType(env)
	var shouldPush *bool
	var hmacInputs, senderHmac *[]byte

	spew.Dump(messageType)
	spew.Dump(topics.V3Conversation)

	if messageType == topics.V2Conversation {
		if parsed, err := parseConversationMessage(env.Message); err == nil {
			shouldPush = parsed.ShouldPush
			hmacInputs = &parsed.HeaderBytes
			if len(parsed.SenderHmac) > 0 {
				senderHmac = &parsed.SenderHmac
			}
		}
	} else {
		fmt.Println("2")
		if message, err := parseGroupMessage(env.Message); err == nil {
			fmt.Println("3")
			push := true
			shouldPush = &push
			hmacInputs = &message.Data
			spew.Dump(message.SenderHmac)
			if len(message.SenderHmac) > 0 {
				fmt.Println("4")
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
