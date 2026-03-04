package xmtp

func getThirtyDayPeriodsFromEpoch(timestamp uint64) int {
	return int(timestamp / 1_000_000_000 / 60 / 60 / 24 / 30)
}
