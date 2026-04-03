package delivery

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	"go.uber.org/zap/zaptest"
)

func newTestRequest() interfaces.SendRequest {
	return interfaces.SendRequest{
		IdempotencyKey: "test-key",
	}
}

// testServerAndDelivery creates an httptest server with the given handler and
// an HttpDelivery pointed at it. The caller should defer server.Close().
func testServerAndDelivery(t *testing.T, handler http.HandlerFunc, maxAttempts int, initialDelayMs int) (*httptest.Server, *HttpDelivery) {
	t.Helper()
	server := httptest.NewServer(handler)
	d := NewHttpDelivery(zaptest.NewLogger(t), options.HttpDeliveryOptions{
		Address:             server.URL,
		MaxAttempts:         maxAttempts,
		InitialRetryDelayMs: initialDelayMs,
	})
	return server, d
}

// countingHandler returns an http.HandlerFunc that counts requests and responds
// with the given status code.
func countingHandler(counter *int32, statusCode int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(counter, 1)
		w.WriteHeader(statusCode)
	}
}

func TestHttpDelivery_SendSuccess(t *testing.T) {
	var requestCount int32
	server, d := testServerAndDelivery(t, countingHandler(&requestCount, http.StatusOK), 3, 10)
	defer server.Close()

	err := d.Send(context.Background(), newTestRequest())
	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&requestCount))
}

func TestHttpDelivery_RetryOnFailureThenSuccess(t *testing.T) {
	var requestCount int32
	server, d := testServerAndDelivery(t, func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		if count == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}, 3, 10)
	defer server.Close()

	err := d.Send(context.Background(), newTestRequest())
	require.NoError(t, err)
	assert.Equal(t, int32(2), atomic.LoadInt32(&requestCount))
}

func TestHttpDelivery_ExhaustsAttempts(t *testing.T) {
	var requestCount int32
	maxAttempts := 3
	server, d := testServerAndDelivery(t, countingHandler(&requestCount, http.StatusInternalServerError), maxAttempts, 10)
	defer server.Close()

	err := d.Send(context.Background(), newTestRequest())
	require.Error(t, err)
	assert.Equal(t, "HTTP request failed", err.Error())
	assert.Equal(t, int32(maxAttempts), atomic.LoadInt32(&requestCount))
}

func TestHttpDelivery_ContextCancellation(t *testing.T) {
	var requestCount int32
	server, d := testServerAndDelivery(t, countingHandler(&requestCount, http.StatusInternalServerError), 5, 500)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- d.Send(ctx, newTestRequest())
	}()

	// Wait for first attempt to complete, then cancel
	time.Sleep(50 * time.Millisecond)
	cancel()

	err := <-done
	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	// Should have made only 1 request before context was cancelled during backoff
	assert.Equal(t, int32(1), atomic.LoadInt32(&requestCount))
}

func TestHttpDelivery_DefaultConfig(t *testing.T) {
	d := NewHttpDelivery(zaptest.NewLogger(t), options.HttpDeliveryOptions{
		Address:             "http://localhost:9999",
		MaxAttempts:         1,
		InitialRetryDelayMs: 250,
	})

	assert.Equal(t, 1, d.maxAttempts)
	assert.Equal(t, 250*time.Millisecond, d.initialRetryDelay)
}

func TestHttpDelivery_ExponentialBackoff(t *testing.T) {
	var timestamps []time.Time
	server, d := testServerAndDelivery(t, func(w http.ResponseWriter, r *http.Request) {
		timestamps = append(timestamps, time.Now())
		w.WriteHeader(http.StatusInternalServerError)
	}, 4, 50)
	defer server.Close()

	_ = d.Send(context.Background(), newTestRequest())

	// Should have 4 requests total (maxAttempts=4)
	require.Len(t, timestamps, 4)

	// Verify delays roughly double each time
	// Expected delays: 50ms, 100ms, 200ms
	for i := 1; i < len(timestamps); i++ {
		gap := timestamps[i].Sub(timestamps[i-1])
		expectedDelay := time.Duration(50*(1<<uint(i-1))) * time.Millisecond
		// Allow 30ms tolerance for test timing
		assert.InDelta(t, expectedDelay.Milliseconds(), gap.Milliseconds(), 30,
			"gap between request %d and %d should be ~%v, got %v", i-1, i, expectedDelay, gap)
	}
}

func TestHttpDelivery_SingleAttempt(t *testing.T) {
	var requestCount int32
	server, d := testServerAndDelivery(t, countingHandler(&requestCount, http.StatusInternalServerError), 1, 10)
	defer server.Close()

	err := d.Send(context.Background(), newTestRequest())
	require.Error(t, err)
	// With maxAttempts=1, only one attempt is made (no retries)
	assert.Equal(t, int32(1), atomic.LoadInt32(&requestCount))
}

func TestHttpDelivery_MaxAttemptsClampsToMinimumOne(t *testing.T) {
	d := NewHttpDelivery(zaptest.NewLogger(t), options.HttpDeliveryOptions{
		Address:             "http://localhost:9999",
		MaxAttempts:         0,
		InitialRetryDelayMs: 10,
	})

	// Value of 0 should be clamped to 1
	assert.Equal(t, 1, d.maxAttempts)
}

func TestHttpDelivery_CanDeliver(t *testing.T) {
	d := NewHttpDelivery(zaptest.NewLogger(t), options.HttpDeliveryOptions{})
	assert.True(t, d.CanDeliver(newTestRequest()))
}

func TestHttpDelivery_AuthHeader(t *testing.T) {
	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	d := NewHttpDelivery(zaptest.NewLogger(t), options.HttpDeliveryOptions{
		Address:             server.URL,
		AuthHeader:          "Bearer test-token",
		MaxAttempts:         1,
		InitialRetryDelayMs: 10,
	})

	err := d.Send(context.Background(), newTestRequest())
	require.NoError(t, err)
	assert.Equal(t, "Bearer test-token", receivedAuth)
}
