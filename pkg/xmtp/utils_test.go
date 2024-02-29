package xmtp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	message_apiv1 "github.com/xmtp/example-notification-server-go/pkg/proto/message_api/v1"
)

func Test_get_thirty_day_periods(t *testing.T) {
	nowNs := time.Now().UnixNano()
	periods := getThirtyDayPeriodsFromEpoch(&message_apiv1.Envelope{
		TimestampNs: uint64(nowNs),
	})
	// This test should cover us for 28 years, but will catch cases where we are off by an order of magnitude
	require.Less(t, periods, 1000)
}
