package xmtp

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/proto/xmtpv4/envelopes"
	"github.com/xmtp/example-notification-server-go/pkg/topics"
)

type messageV4Info struct {
	context interfaces.MessageContext

	originatorNs    int64
	idempotencyData *[]byte
}

func parseV4Envelope(env *envelopes.OriginatorEnvelope) (*messageV4Info, bool, error) {

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
	if groupMessage == nil {
		// Not a group message - nothing else to be done.
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

	// Essentially maintaining compatibility with the previous classification.
	switch topicKind {

	case topics.TopicKindGroupMessagesV1:
		messageType = topics.V3Conversation

	case topics.TopicKindWelcomeMessagesV1:
		messageType = topics.V3Welcome

	default:
		return nil, false, nil
	}

	var (
		senderHmac  = groupMessage.GetV1().GetSenderHmac()
		shouldPush  = groupMessage.GetV1().GetShouldPush()
		messageData = groupMessage.GetV1().GetData()
	)

	// TODO: Check - is it the V1().GetData() that is the new equivalent of the old data, or is it the unsigned envelope payload
	out := messageV4Info{
		context: interfaces.MessageContext{
			MessageType: messageType,
			SenderHmac:  &senderHmac,
			ShouldPush:  &shouldPush,
			Topic:       topic.String(),
			HmacInputs:  &messageData,
		},

		originatorNs:    unsignedEnv.OriginatorNs,
		idempotencyData: &messageData,
	}

	return &out, true, nil
}
