package xmtp

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xmtp/xmtpd/pkg/envelopes"
	messageApi "github.com/xmtp/xmtpd/pkg/proto/xmtpv4/message_api"
)

func TestXmtpdNotificationApiImportable(t *testing.T) {
	// This test verifies the xmtpd notification API proto types are available.
	// It will fail until xmtpd is pinned to the notification API branch.
	var _ messageApi.NotificationApiClient
}

func TestXmtpdEnvelopeTypesImportable(t *testing.T) {
	// Verify that importing xmtpd/pkg/envelopes does NOT cause
	// protobuf namespace conflicts after local proto dirs are deleted.
	_, err := envelopes.NewOriginatorEnvelopeFromBytes([]byte{})
	require.Error(t, err) // Expected: parse error on empty bytes, not a panic
}
