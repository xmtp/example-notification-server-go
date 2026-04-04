package xmtp

import (
	"github.com/xmtp/xmtpd/pkg/envelopes"
	mlsV1 "github.com/xmtp/xmtpd/pkg/proto/mls/api/v1"
	"github.com/xmtp/xmtpd/pkg/topic"
	"google.golang.org/protobuf/proto"
)

// cloneBytes returns a defensive copy of a byte slice. Returns nil if input is nil.
func cloneBytes(input []byte) []byte {
	if input == nil {
		return nil
	}
	return append([]byte(nil), input...)
}

// convertGroupMessageToV3 converts a V4 GroupMessageInput_V1 and its originator envelope
// into a serialized V3-format GroupMessage.
func convertGroupMessageToV3(
	input *mlsV1.GroupMessageInput_V1,
	origEnv *envelopes.OriginatorEnvelope,
	targetTopic *topic.Topic,
) ([]byte, error) {
	msg := &mlsV1.GroupMessage{
		Version: &mlsV1.GroupMessage_V1_{
			V1: &mlsV1.GroupMessage_V1{
				Id:         origEnv.OriginatorSequenceID(),
				CreatedNs:  uint64(origEnv.OriginatorNs()),
				GroupId:    cloneBytes(targetTopic.Identifier()),
				Data:       cloneBytes(input.Data),
				SenderHmac: cloneBytes(input.SenderHmac),
				ShouldPush: input.ShouldPush,
				IsCommit:   origEnv.OriginatorNodeID() == 0,
			},
		},
	}
	return proto.Marshal(msg)
}

// convertWelcomeMessageToV3 converts a V4 WelcomeMessageInput_V1 and its originator envelope
// into a serialized V3-format WelcomeMessage.
func convertWelcomeMessageToV3(
	input *mlsV1.WelcomeMessageInput_V1,
	origEnv *envelopes.OriginatorEnvelope,
) ([]byte, error) {
	msg := &mlsV1.WelcomeMessage{
		Version: &mlsV1.WelcomeMessage_V1_{
			V1: &mlsV1.WelcomeMessage_V1{
				Id:               origEnv.OriginatorSequenceID(),
				CreatedNs:        uint64(origEnv.OriginatorNs()),
				InstallationKey:  cloneBytes(input.InstallationKey),
				Data:             cloneBytes(input.Data),
				HpkePublicKey:    cloneBytes(input.HpkePublicKey),
				WrapperAlgorithm: input.WrapperAlgorithm,
				WelcomeMetadata:  cloneBytes(input.WelcomeMetadata),
			},
		},
	}
	return proto.Marshal(msg)
}

// convertWelcomePointerToV3 converts a V4 WelcomeMessageInput_WelcomePointer and its originator
// envelope into a serialized V3-format WelcomeMessage with WelcomePointer variant.
func convertWelcomePointerToV3(
	input *mlsV1.WelcomeMessageInput_WelcomePointer,
	origEnv *envelopes.OriginatorEnvelope,
) ([]byte, error) {
	msg := &mlsV1.WelcomeMessage{
		Version: &mlsV1.WelcomeMessage_WelcomePointer_{
			WelcomePointer: &mlsV1.WelcomeMessage_WelcomePointer{
				Id:               origEnv.OriginatorSequenceID(),
				CreatedNs:        uint64(origEnv.OriginatorNs()),
				InstallationKey:  cloneBytes(input.InstallationKey),
				WelcomePointer:   cloneBytes(input.WelcomePointer),
				HpkePublicKey:    cloneBytes(input.HpkePublicKey),
				WrapperAlgorithm: input.WrapperAlgorithm,
			},
		},
	}
	return proto.Marshal(msg)
}
