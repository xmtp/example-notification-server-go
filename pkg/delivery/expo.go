package delivery

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	"go.uber.org/zap"
)

const expoAPIURL = "https://exp.host/--/api/v2/push/send?useFcmV1=true"

type ExpoDelivery struct {
	logger      *zap.Logger
	httpClient  *http.Client
	accessToken string
}

type expoPushMessage struct {
	To       string            `json:"to"`
	Title    string            `json:"title,omitempty"`
	Body     string            `json:"body,omitempty"`
	Data     map[string]string `json:"data"`
	Priority string            `json:"priority,omitempty"`
	Sound    string            `json:"sound,omitempty"`
	Badge    int               `json:"badge"`
}

type expoResponse struct {
	Data []struct {
		Status  string `json:"status"`
		Message string `json:"message,omitempty"`
		Details struct {
			Error string `json:"error,omitempty"`
		} `json:"details,omitempty"`
	} `json:"data"`
}

func NewExpoDelivery(logger *zap.Logger, opts options.ExpoOptions) *ExpoDelivery {
	return &ExpoDelivery{
		logger:      logger,
		accessToken: opts.AccessToken,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (e *ExpoDelivery) CanDeliver(req interfaces.SendRequest) bool {
	if req.Installation.DeliveryMechanism.Kind != interfaces.EXPO {
		return false
	}
	token := req.Installation.DeliveryMechanism.Token
	if token == "" {
		return false
	}
	// Validate Expo token format
	return strings.HasPrefix(token, "ExponentPushToken[") || strings.HasPrefix(token, "ExpoPushToken[")
}

func (e *ExpoDelivery) Send(ctx context.Context, req interfaces.SendRequest) error {
	if req.Message == nil {
		return errors.New("missing message")
	}

	message := base64.StdEncoding.EncodeToString(req.Message.Message)
	topic := req.Message.ContentTopic

	data := map[string]string{
		"topic":            topic,
		"encryptedMessage": message,
		"messageType":      string(req.MessageContext.MessageType),
	}

	// Determine priority and notification fields based on silent mode
	priority := "high"
	title := ""
	body := ""
	sound := ""

	if !req.Subscription.IsSilent {
		title = "New XMTP Message"
		body = "You have a new message"
		sound = "default"
	} else {
		priority = "normal"
	}

	pushMessage := expoPushMessage{
		To:       req.Installation.DeliveryMechanism.Token,
		Title:    title,
		Body:     body,
		Data:     data,
		Priority: priority,
		Sound:    sound,
		Badge:    0,
	}

	// Expo API expects an array of messages
	messages := []expoPushMessage{pushMessage}
	payload, err := json.Marshal(messages)
	if err != nil {
		return errors.Wrap(err, "failed to marshal expo push message")
	}

	e.logger.Info("sending expo push notification",
		zap.String("device_token", req.Installation.DeliveryMechanism.Token),
		zap.String("topic", topic),
		zap.Bool("has_access_token", e.accessToken != ""),
	)

	// Retry logic with exponential backoff
	maxRetries := 3
	var resp *http.Response
	var bodyBytes []byte

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
			e.logger.Info("retrying expo push notification",
				zap.Int("attempt", attempt+1),
				zap.Duration("backoff", backoff))

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", expoAPIURL, bytes.NewBuffer(payload))
		if err != nil {
			return errors.Wrap(err, "failed to create http request")
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Accept", "application/json")
		httpReq.Header.Set("Accept-Encoding", "gzip, deflate")

		// Add Authorization header if access token is provided
		if e.accessToken != "" {
			httpReq.Header.Set("Authorization", "Bearer "+e.accessToken)
		}

		resp, err = e.httpClient.Do(httpReq)
		if err != nil {
			if attempt == maxRetries-1 {
				return errors.Wrap(err, "failed to send expo push notification after retries")
			}
			e.logger.Warn("expo push notification request failed, will retry",
				zap.Error(err),
				zap.Int("attempt", attempt+1))
			continue
		}

		// Read response body
		bodyBytes, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			if attempt == maxRetries-1 {
				return errors.Wrap(err, "failed to read response body after retries")
			}
			e.logger.Warn("failed to read response body, will retry",
				zap.Error(err),
				zap.Int("attempt", attempt+1))
			continue
		}

		// Retry on 5xx errors or 429 (rate limit)
		if resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests {
			if attempt == maxRetries-1 {
				e.logger.Error("Expo API error response after retries",
					zap.Int("status", resp.StatusCode),
					zap.String("body", string(bodyBytes)),
					zap.String("token", req.Installation.DeliveryMechanism.Token))
				return fmt.Errorf("expo API returned status %d after retries: %s", resp.StatusCode, string(bodyBytes))
			}
			e.logger.Warn("expo API returned retryable error, will retry",
				zap.Int("status", resp.StatusCode),
				zap.Int("attempt", attempt+1))
			continue
		}

		// Success or non-retryable error, break out of retry loop
		break
	}

	if resp.StatusCode != http.StatusOK {
		// Log error response for debugging
		e.logger.Error("Expo API error response",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(bodyBytes)),
			zap.String("token", req.Installation.DeliveryMechanism.Token))
		return fmt.Errorf("expo API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var expoResp expoResponse
	if err := json.Unmarshal(bodyBytes, &expoResp); err != nil {
		return errors.Wrap(err, "failed to decode expo response")
	}

	// Check all response items for errors
	for i, result := range expoResp.Data {
		if result.Status == "error" {
			errType := result.Details.Error
			errMsg := result.Message
			if errType != "" {
				errMsg = errType
			}

			e.logger.Error("expo push failed",
				zap.Int("index", i),
				zap.String("error_type", errType),
				zap.String("error_message", result.Message),
				zap.String("token", req.Installation.DeliveryMechanism.Token))

			// Handle specific error types
			switch errType {
			case "DeviceNotRegistered":
				// Token is invalid/expired - should trigger cleanup in the future
				return fmt.Errorf("expo push failed: device not registered (invalid token)")
			case "MessageTooBig":
				return fmt.Errorf("expo push failed: message too big")
			case "MessageRateExceeded":
				return fmt.Errorf("expo push failed: message rate exceeded")
			case "InvalidCredentials":
				return fmt.Errorf("expo push failed: invalid credentials")
			default:
				return fmt.Errorf("expo push failed: %s", errMsg)
			}
		}
	}

	e.logger.Info("expo push notification sent successfully",
		zap.String("token", req.Installation.DeliveryMechanism.Token),
		zap.String("topic", topic),
	)

	return nil
}
