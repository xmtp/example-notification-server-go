package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	"go.uber.org/zap"
)

type HttpDelivery struct {
	address    string
	authHeader string
	logger     *zap.Logger
}

func NewHttpDelivery(logger *zap.Logger, opts options.HttpDeliveryOptions) *HttpDelivery {
	return &HttpDelivery{
		logger:     logger,
		address:    opts.Address,
		authHeader: opts.AuthHeader,
	}
}

func (h HttpDelivery) CanDeliver(req interfaces.SendRequest) bool {
	return true
}

func (h HttpDelivery) Send(ctx context.Context, req interfaces.SendRequest) error {
	// Convert the request data to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return err
	}

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
