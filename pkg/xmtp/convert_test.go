package xmtp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xmtp/xmtpd/pkg/envelopes"
	mlsV1 "github.com/xmtp/xmtpd/pkg/proto/mls/api/v1"
	testEnvelopes "github.com/xmtp/xmtpd/pkg/testutils/envelopes"
	"github.com/xmtp/xmtpd/pkg/topic"
	"google.golang.org/protobuf/proto"
)

func buildTestOriginatorEnvelope(
	t *testing.T,
	nodeID uint32,
	sequenceID uint64,
	timestampNs int64,
) *envelopes.OriginatorEnvelope {
	t.Helper()
	protoEnv := testEnvelopes.CreateOriginatorEnvelopeWithTimestamp(
		t,
		nodeID,
		sequenceID,
		time.Unix(0, timestampNs),
	)
	origEnv, err := envelopes.NewOriginatorEnvelope(protoEnv)
	require.NoError(t, err)
	return origEnv
}

func TestConvertGroupMessageToV3(t *testing.T) {
	const (
		nodeID     = uint32(0) // IsCommit = true when nodeID == 0
		sequenceID = uint64(42)
		tsNs       = int64(1_000_000_000)
	)

	groupID := []byte{0x01, 0x02, 0x03, 0x04}
	targetTopic := topic.NewTopic(topic.TopicKindGroupMessagesV1, groupID)

	input := &mlsV1.GroupMessageInput_V1{
		Data:       []byte("group-data"),
		SenderHmac: []byte("hmac-value"),
		ShouldPush: true,
	}

	origEnv := buildTestOriginatorEnvelope(t, nodeID, sequenceID, tsNs)

	result, err := convertGroupMessageToV3(input, origEnv, targetTopic)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	var msg mlsV1.GroupMessage
	err = proto.Unmarshal(result, &msg)
	require.NoError(t, err)

	v1 := msg.GetV1()
	require.NotNil(t, v1)

	require.Equal(t, sequenceID, v1.GetId())
	require.Equal(t, uint64(tsNs), v1.GetCreatedNs())
	require.Equal(t, groupID, v1.GetGroupId())
	require.Equal(t, input.Data, v1.GetData())
	require.Equal(t, input.SenderHmac, v1.GetSenderHmac())
	require.Equal(t, input.ShouldPush, v1.GetShouldPush())
	require.True(t, v1.GetIsCommit(), "IsCommit should be true when nodeID == 0")
}

func TestConvertGroupMessageToV3_NotCommit(t *testing.T) {
	const (
		nodeID     = uint32(5) // IsCommit = false when nodeID != 0
		sequenceID = uint64(7)
		tsNs       = int64(2_000_000_000)
	)

	groupID := []byte{0xAB, 0xCD}
	targetTopic := topic.NewTopic(topic.TopicKindGroupMessagesV1, groupID)

	input := &mlsV1.GroupMessageInput_V1{
		Data:       []byte("not-a-commit"),
		SenderHmac: []byte("some-hmac"),
		ShouldPush: false,
	}

	origEnv := buildTestOriginatorEnvelope(t, nodeID, sequenceID, tsNs)

	result, err := convertGroupMessageToV3(input, origEnv, targetTopic)
	require.NoError(t, err)

	var msg mlsV1.GroupMessage
	err = proto.Unmarshal(result, &msg)
	require.NoError(t, err)

	v1 := msg.GetV1()
	require.NotNil(t, v1)

	require.Equal(t, sequenceID, v1.GetId())
	require.Equal(t, uint64(tsNs), v1.GetCreatedNs())
	require.False(t, v1.GetIsCommit(), "IsCommit should be false when nodeID != 0")
}

func TestConvertWelcomePointerToV3(t *testing.T) {
	const (
		nodeID     = uint32(2)
		sequenceID = uint64(55)
		tsNs       = int64(4_000_000_000)
	)

	input := &mlsV1.WelcomeMessageInput_WelcomePointer{
		InstallationKey:  []byte("install-key"),
		WelcomePointer:   []byte("welcome-ptr-data"),
		HpkePublicKey:    []byte("hpke-key"),
		WrapperAlgorithm: 0,
	}

	origEnv := buildTestOriginatorEnvelope(t, nodeID, sequenceID, tsNs)

	result, err := convertWelcomePointerToV3(input, origEnv)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	var msg mlsV1.WelcomeMessage
	err = proto.Unmarshal(result, &msg)
	require.NoError(t, err)

	wp := msg.GetWelcomePointer()
	require.NotNil(t, wp)
	require.Equal(t, sequenceID, wp.GetId())
	require.Equal(t, uint64(tsNs), wp.GetCreatedNs())
	require.Equal(t, input.InstallationKey, wp.GetInstallationKey())
	require.Equal(t, input.WelcomePointer, wp.GetWelcomePointer())
	require.Equal(t, input.HpkePublicKey, wp.GetHpkePublicKey())
}

func TestConvertWelcomeMessageToV3(t *testing.T) {
	const (
		nodeID     = uint32(1)
		sequenceID = uint64(99)
		tsNs       = int64(3_000_000_000)
	)

	input := &mlsV1.WelcomeMessageInput_V1{
		InstallationKey:  []byte("installation-key"),
		Data:             []byte("welcome-data"),
		HpkePublicKey:    []byte("hpke-key"),
		WrapperAlgorithm: 0,
		WelcomeMetadata:  []byte("metadata"),
	}

	origEnv := buildTestOriginatorEnvelope(t, nodeID, sequenceID, tsNs)

	result, err := convertWelcomeMessageToV3(input, origEnv)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	var msg mlsV1.WelcomeMessage
	err = proto.Unmarshal(result, &msg)
	require.NoError(t, err)

	v1 := msg.GetV1()
	require.NotNil(t, v1)

	require.Equal(t, sequenceID, v1.GetId())
	require.Equal(t, uint64(tsNs), v1.GetCreatedNs())
	require.Equal(t, input.InstallationKey, v1.GetInstallationKey())
	require.Equal(t, input.Data, v1.GetData())
	require.Equal(t, input.HpkePublicKey, v1.GetHpkePublicKey())
	require.Equal(t, input.WrapperAlgorithm, v1.GetWrapperAlgorithm())
	require.Equal(t, input.WelcomeMetadata, v1.GetWelcomeMetadata())
}
