package testutils

import (
	"testing"

	"github.com/stretchr/testify/require"
	topicutil "github.com/xmtp/example-notification-server-go/pkg/topics"
	topicpkg "github.com/xmtp/xmtpd/pkg/topic"
)

// MustParseTopic parses a V3 topic string or fails the test.
func MustParseTopic(t *testing.T, topicStr string) *topicpkg.Topic {
	t.Helper()
	parsed, err := topicutil.ParseV3Topic(topicStr)
	require.NoError(t, err)
	return parsed
}

// MustParseTopicBytes parses a V3 topic string and returns raw bytes, or fails.
func MustParseTopicBytes(t *testing.T, topicStr string) []byte {
	t.Helper()
	return MustParseTopic(t, topicStr).Bytes()
}
