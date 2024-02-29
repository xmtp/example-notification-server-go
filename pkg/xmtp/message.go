package xmtp

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"

	messageApi "github.com/xmtp/example-notification-server-go/pkg/proto/message_api/v1"
	messageContents "github.com/xmtp/example-notification-server-go/pkg/proto/message_contents"
	"google.golang.org/protobuf/proto"
)

type MessageContext struct {
	MessageType MessageType
	ShouldPush  *bool
	IsSender    *bool
}

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

func getIsSender(msg *messageContents.MessageV2, hmacKey *[]byte) *bool {
	isSender := false
	if len(msg.SenderHmac) > 0 && hmacKey != nil {
		fmt.Printf("Got HMAC key %x and sender hmac %x", hmacKey, msg.SenderHmac)
		// Calculate HMAC of the HeaderBytes using the provided key and compare it with the SenderHmac
		hmacHash := hmac.New(sha256.New, *hmacKey)
		hmacHash.Write(msg.HeaderBytes)
		expectedHmac := hmacHash.Sum(nil)
		isSender = hmac.Equal(msg.SenderHmac, expectedHmac)
	}
	return &isSender
}

func getContext(env *messageApi.Envelope, hmacKey *[]byte) MessageContext {
	messageType := getMessageType(env)
	var shouldPush, isSender *bool
	if messageType == V2Conversation {
		if parsed, err := parseConversationMessage(env.Message); err == nil {
			shouldPush = parsed.ShouldPush
			isSender = getIsSender(parsed, hmacKey)
		}
	}

	return MessageContext{
		MessageType: messageType,
		ShouldPush:  shouldPush,
		IsSender:    isSender,
	}
}
