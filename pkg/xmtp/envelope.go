package xmtp

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/proto/xmtpv4/envelopes"
	"github.com/xmtp/example-notification-server-go/pkg/topics"
)

func parseV4Envelope(env *envelopes.OriginatorEnvelope) (*interfaces.MessageContext, bool, error) {

	ue := env.GetUnsignedOriginatorEnvelope()
	if ue == nil {
		return nil, false, errors.New("unsigned originator envelope missing")
	}

	var unsignedEnv envelopes.UnsignedOriginatorEnvelope
	err := proto.Unmarshal(ue, &unsignedEnv)
	if err != nil {
		return nil, false, fmt.Errorf("could not decode unsigned envelope: %w", err)
	}

	pe := unsignedEnv.GetPayerEnvelopeBytes()
	if pe == nil {
		return nil, false, errors.New("payer envelope missing")
	}

	var payerEnv envelopes.PayerEnvelope
	err = proto.Unmarshal(pe, &payerEnv)
	if err != nil {
		return nil, false, fmt.Errorf("could not decode payer envelope: %w", err)
	}

	ce := payerEnv.GetUnsignedClientEnvelope()
	if ce == nil {
		return nil, false, errors.New("client envelope missing")
	}

	var clientEnv envelopes.ClientEnvelope
	err = proto.Unmarshal(ce, &clientEnv)
	if err != nil {
		return nil, false, fmt.Errorf("could not decode client envelope: %w", err)
	}

	groupMessage := clientEnv.GetGroupMessage()

	// Not a group message - nothing else to be done.
	if groupMessage == nil {
		return nil, false, nil
	}

	topic, err := topics.ParseTopic(clientEnv.GetAad().GetTargetTopic())
	if err != nil {
		return nil, false, fmt.Errorf("could not parse topic: %w", err)
	}

	// Determine the message type.
	var (
		topicKind   = topic.Kind()
		messageType topics.MessageType
	)

	if topicKind == topics.TopicKindGroupMessagesV1 {
		messageType = topics.V3Conversation
	} else if topicKind == topics.TopicKindWelcomeMessagesV1 {
		messageType = topics.V3Welcome
	} else {
		return nil, false, nil
	}

	fmt.Printf(">>> topic - kind: %v, reserved: %v, str: %v\n", topicKind, topic.IsReserved(), topic.String())

	var (
		senderHmac  = groupMessage.GetV1().GetSenderHmac()
		shouldPush  = groupMessage.GetV1().GetShouldPush()
		messageData = groupMessage.GetV1().GetData()
	)

	// TODO: Handle timestamp

	out := interfaces.MessageContext{
		MessageType: messageType,
		SenderHmac:  &senderHmac,
		ShouldPush:  &shouldPush,
		Topic:       topic.String(),
		HmacInputs:  &messageData,
	}

	return &out, true, nil
}
