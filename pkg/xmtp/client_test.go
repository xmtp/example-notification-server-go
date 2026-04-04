package xmtp

import (
	"testing"

	messageApi "github.com/xmtp/xmtpd/pkg/proto/xmtpv4/message_api"
)

func TestXmtpdNotificationApiImportable(t *testing.T) {
	// This test verifies the xmtpd notification API proto types are available.
	// It will fail until xmtpd is pinned to the notification API branch.
	var _ messageApi.NotificationApiClient
}
