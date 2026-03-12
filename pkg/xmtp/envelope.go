package xmtp

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	apiv1 "github.com/xmtp/example-notification-server-go/pkg/proto/mls/api/v1"
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

	// Determine what kind of message this is.
	topic, err := topics.ParseTopic(clientEnv.GetAad().GetTargetTopic())
	if err != nil {
		return nil, false, fmt.Errorf("could not parse topic: %w", err)
	}

	switch topic.Kind() {

	case topics.TopicKindWelcomeMessagesV1:
		return parseV4WelcomeMessage(clientEnv.GetWelcomeMessage(), topic.String(), unsignedEnv.GetOriginatorNs())

	case topics.TopicKindGroupMessagesV1:
		return parseV4GroupMessage(clientEnv.GetGroupMessage(), topic.String(), unsignedEnv.GetOriginatorNs())

	default:
		return nil, false, nil
	}
}

func parseV4GroupMessage(groupMessage *apiv1.GroupMessageInput, topic string, originatorNs int64) (*messageV4Info, bool, error) {

	if groupMessage == nil {
		// Should not happen as topic kind told us what it was
		return nil, false, errors.New("group message missing")
	}

	var (
		senderHmac  = groupMessage.GetV1().GetSenderHmac()
		shouldPush  = groupMessage.GetV1().GetShouldPush()
		messageData = groupMessage.GetV1().GetData()
	)

	// We use the V1 Data as inputs for both idempotency key and HMAC input, equivalent to the old behavior.
	out := messageV4Info{
		context: interfaces.MessageContext{
			// Essentially maintaining compatibility with the previous classification.
			// TODO: Check - might include V4 variants?
			MessageType: topics.V3Conversation,
			SenderHmac:  &senderHmac,
			ShouldPush:  &shouldPush,
			Topic:       topic,

			// Right now separated to make it clear what is HmacInput and what is message payload.
			// But instead we could be transparent, just have message payload and use it for HMAC / idempotency key too.
			HmacInputs:       &messageData,
			MessagePayloadV4: messageData,
		},

		originatorNs:    originatorNs,
		idempotencyData: &messageData,
	}

	return &out, true, nil
}

func parseV4WelcomeMessage(welcomeMessage *apiv1.WelcomeMessageInput, topic string, originatorNs int64) (*messageV4Info, bool, error) {
	if welcomeMessage == nil {
		// Should not happen as topic kind told us what it was
		return nil, false, errors.New("welcome message missing")
	}

	messageData := welcomeMessage.GetV1().GetData()

	// We use the V1 Data as inputs for both idempotency key and HMAC input, equivalent to the old behavior.
	out := messageV4Info{
		context: interfaces.MessageContext{
			// Essentially maintaining compatibility with the previous classification.
			// TODO: Check - might include V4 variants?
			MessageType:      topics.V3Welcome,
			Topic:            topic,
			MessagePayloadV4: messageData,
		},
		originatorNs:    originatorNs,
		idempotencyData: &messageData,
	}

	return &out, true, nil
}
