package testutils

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/xmtp/example-notification-server-go/mocks"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
)

// MockDeliveryWithSendCounter creates a mock Delivery that accepts all sends,
// increments the counter on each Send call, and returns nil.
func MockDeliveryWithSendCounter(t *testing.T) (*mocks.Delivery, *atomic.Int32) {
	t.Helper()
	m := mocks.NewDelivery(t)
	m.On("CanDeliver", mock.Anything).Return(true)
	var count atomic.Int32
	m.On("Send", mock.Anything, mock.Anything).
		Run(func(mock.Arguments) { count.Add(1) }).
		Return(nil)
	return m, &count
}

// MockDeliveryAcceptAll creates a mock Delivery that accepts and succeeds on all sends.
func MockDeliveryAcceptAll(t *testing.T) *mocks.Delivery {
	t.Helper()
	m := mocks.NewDelivery(t)
	m.On("CanDeliver", mock.Anything).Return(true)
	m.On("Send", mock.Anything, mock.Anything).Return(nil)
	return m
}

// RequireEventuallySendCount asserts that the counter reaches want within 1 second.
func RequireEventuallySendCount(t *testing.T, counter *atomic.Int32, want int32) {
	t.Helper()
	require.Eventually(t, func() bool {
		return counter.Load() == want
	}, time.Second, 10*time.Millisecond)
}

// GetSendRequests extracts all SendRequest arguments from a mock Delivery's Send calls.
func GetSendRequests(m *mocks.Delivery) []interfaces.SendRequest {
	var reqs []interfaces.SendRequest
	for _, call := range m.Calls {
		if call.Method == "Send" {
			reqs = append(reqs, call.Arguments.Get(1).(interfaces.SendRequest))
		}
	}
	return reqs
}

// RequireSendRequestForInstallation finds the SendRequest for a given installation ID or fails.
func RequireSendRequestForInstallation(t *testing.T, reqs []interfaces.SendRequest, installationID string) interfaces.SendRequest {
	t.Helper()
	for _, req := range reqs {
		if req.Installation.Id == installationID {
			return req
		}
	}
	require.Failf(t, "missing send request", "installation %q was not delivered", installationID)
	return interfaces.SendRequest{}
}
