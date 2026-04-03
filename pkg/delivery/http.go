package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	"go.uber.org/zap"
)

type HttpDelivery struct {
	address           string
	authHeader        string
	logger            *zap.Logger
	maxAttempts       int
	initialRetryDelay time.Duration
}

func NewHttpDelivery(logger *zap.Logger, opts options.HttpDeliveryOptions) *HttpDelivery {
	maxAttempts := opts.MaxAttempts
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	return &HttpDelivery{
		logger:            logger,
		address:           opts.Address,
		authHeader:        opts.AuthHeader,
		maxAttempts:       maxAttempts,
		initialRetryDelay: time.Duration(opts.InitialRetryDelayMs) * time.Millisecond,
	}
}

func (h HttpDelivery) CanDeliver(req interfaces.SendRequest) bool {
	return true
}

func (h HttpDelivery) Send(ctx context.Context, req interfaces.SendRequest) error {
	// Convert the request data to JSON (non-retryable)
	jsonData, err := json.Marshal(req)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt < h.maxAttempts; attempt++ {
		lastErr = h.doSend(ctx, jsonData)
		if lastErr == nil {
			return nil
		}

		if attempt < h.maxAttempts-1 {
			delay := h.initialRetryDelay * (1 << uint(attempt))
			h.logger.Warn("HTTP delivery failed, retrying",
				zap.Int("attempt", attempt+1),
				zap.Int("max_attempts", h.maxAttempts),
				zap.Duration("next_delay", delay),
				zap.Error(lastErr),
			)
			select {
			case <-time.After(delay):
				// continue to next attempt
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return lastErr
}

func (h HttpDelivery) doSend(ctx context.Context, jsonData []byte) error {
	// Create a new HTTP request with context
	httpRequest, err := http.NewRequestWithContext(ctx, "POST", h.address, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	// Set the content type and authorization headers
	httpRequest.Header.Set("Content-Type", "application/json")
	if h.authHeader != "" {
		httpRequest.Header.Set("Authorization", h.authHeader)
	}

	// Send the request using the http.DefaultClient
	response, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	// Check the response status code
	if response.StatusCode != http.StatusOK {
		h.logger.Error("HTTP request failed",
			zap.Int("status_code", response.StatusCode),
		)
		return errors.New("HTTP request failed")
	}

	return nil
}
