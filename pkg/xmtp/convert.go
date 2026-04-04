package xmtp

import (
	"github.com/xmtp/xmtpd/pkg/envelopes"
	mlsV1 "github.com/xmtp/xmtpd/pkg/proto/mls/api/v1"
	"github.com/xmtp/xmtpd/pkg/topic"
	"google.golang.org/protobuf/proto"
)

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
				GroupId:    targetTopic.Identifier(),
				Data:       input.Data,
				SenderHmac: input.SenderHmac,
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
				InstallationKey:  input.InstallationKey,
				Data:             input.Data,
				HpkePublicKey:    input.HpkePublicKey,
				WrapperAlgorithm: input.WrapperAlgorithm,
				WelcomeMetadata:  input.WelcomeMetadata,
			},
		},
	}
	return proto.Marshal(msg)
}
