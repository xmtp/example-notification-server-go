package topics

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xmtp/xmtpd/pkg/topic"
)

func TestParseV3Topic(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantKind       topic.TopicKind
		wantIdentifier string // hex-encoded expected identifier (empty if wantErr)
		wantErr        bool
	}{
		{
			name:           "group message",
			input:          "/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto",
			wantKind:       topic.TopicKindGroupMessagesV1,
			wantIdentifier: "24ce39d660600b3a98adff3075b6d1f4",
		},
		{
			name:     "welcome message",
			input:    "/xmtp/mls/1/w-f3ac64eba2272334124975d673374bdd64a5535bf5f7b48ac5608ff499444be0/proto",
			wantKind: topic.TopicKindWelcomeMessagesV1,
		},
		{
			name:     "mixed case hex",
			input:    "/xmtp/mls/1/g-24CE39D660600B3A98ADFF3075B6D1F4/proto",
			wantKind: topic.TopicKindGroupMessagesV1,
		},
		{
			name:    "invalid pattern",
			input:   "/xmtp/mls/1/x-abc123/proto",
			wantErr: true,
		},
		{
			name:    "not hex",
			input:   "/xmtp/mls/1/g-zzzz/proto",
			wantErr: true,
		},
		{
			name:    "missing suffix",
			input:   "/xmtp/mls/1/g-24ce39d660600b3a",
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseV3Topic(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.wantKind, result.Kind())
			if tc.wantIdentifier != "" {
				expected, _ := hex.DecodeString(tc.wantIdentifier)
				require.Equal(t, expected, result.Identifier())
			}
		})
	}
}

func TestTopicToString(t *testing.T) {
	tests := []struct {
		name       string
		kind       topic.TopicKind
		identifier string // hex-encoded
		want       string
	}{
		{
			name:       "group message",
			kind:       topic.TopicKindGroupMessagesV1,
			identifier: "24ce39d660600b3a98adff3075b6d1f4",
			want:       "/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto",
		},
		{
			name:       "welcome message",
			kind:       topic.TopicKindWelcomeMessagesV1,
			identifier: "f3ac64eba2272334",
			want:       "/xmtp/mls/1/w-f3ac64eba2272334/proto",
		},
		{
			name:       "unknown kind returns empty",
			kind:       topic.TopicKindIdentityUpdatesV1,
			identifier: "abcd",
			want:       "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			identifier, _ := hex.DecodeString(tc.identifier)
			tp := topic.NewTopic(tc.kind, identifier)
			require.Equal(t, tc.want, TopicToString(tp))
		})
	}
}

func TestTopicToString_Roundtrip(t *testing.T) {
	original := "/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto"
	parsed, err := ParseV3Topic(original)
	require.NoError(t, err)
	require.Equal(t, original, TopicToString(parsed))
}

func TestGetMessageTypeFromTopic(t *testing.T) {
	tests := []struct {
		name  string
		topic *topic.Topic
		want  MessageType
	}{
		{
			name:  "group",
			topic: topic.NewTopic(topic.TopicKindGroupMessagesV1, []byte{0x01}),
			want:  V3Conversation,
		},
		{
			name:  "welcome",
			topic: topic.NewTopic(topic.TopicKindWelcomeMessagesV1, []byte{0x01}),
			want:  V3Welcome,
		},
		{
			name:  "unknown kind",
			topic: topic.NewTopic(topic.TopicKindIdentityUpdatesV1, []byte{0x01}),
			want:  Unknown,
		},
		{
			name:  "nil topic",
			topic: nil,
			want:  Unknown,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, GetMessageTypeFromTopic(tc.topic))
		})
	}
}
