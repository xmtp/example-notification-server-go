package xmtp

import (
	v1 "github.com/xmtp/example-notification-server-go/pkg/proto/message_api/v1"
)

func getThirtyDayPeriodsFromEpoch(env *v1.Envelope) int {
	return int(env.TimestampNs / 1_000_000_000 / 60 / 60 / 24 / 30)
}
