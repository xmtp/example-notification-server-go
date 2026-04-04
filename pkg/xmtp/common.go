package xmtp

import (
	"context"
	"time"

	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"go.uber.org/zap"
)

const STARTING_SLEEP_TIME = 100 * time.Millisecond
const MAX_SLEEP_TIME = 30 * time.Second
const DELIVERY_TIMEOUT = 15 * time.Second

// cappedBackoff doubles sleepTime up to MAX_SLEEP_TIME.
func cappedBackoff(sleepTime time.Duration) time.Duration {
	sleepTime *= 2
	if sleepTime > MAX_SLEEP_TIME {
		sleepTime = MAX_SLEEP_TIME
	}
	return sleepTime
}

// NotificationListener is the interface implemented by both V3 and V4 listeners
type NotificationListener interface {
	Start()
	Stop()
}

// deliveryDispatcher handles shared delivery logic for both V3 and V4 listeners
type deliveryDispatcher struct {
	logger           *zap.Logger
	ctx              context.Context
	deliveryServices []interfaces.Delivery
}

func (d *deliveryDispatcher) shouldDeliver(messageContext interfaces.MessageContext, subscription interfaces.Subscription) bool {
	if subscription.HmacKey != nil && len(subscription.HmacKey.Key) > 0 {
		isSender := messageContext.IsSender(subscription.HmacKey.Key)
		if isSender {
			return false
		}
	}
	if messageContext.ShouldPush != nil {
		shouldPush := messageContext.ShouldPush
		return *shouldPush
	}
	return true
}

func (d *deliveryDispatcher) deliver(req interfaces.SendRequest) error {
	ctx, cancel := context.WithTimeout(d.ctx, DELIVERY_TIMEOUT)
	defer cancel()
	for _, service := range d.deliveryServices {
		if service.CanDeliver(req) {
			d.logger.Info("active subscription found. sending message",
				zap.String("topic", req.Topic),
				zap.String("message_type", string(req.MessageContext.MessageType)),
			)
			return service.Send(ctx, req)
		}
	}
	d.logger.Info("No delivery service matches request", zap.String("delivery_mechanism", string(req.Installation.DeliveryMechanism.Kind)))
	return nil
}
