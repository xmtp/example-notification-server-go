package delivery

import (
	"context"

	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"go.uber.org/zap"
)

type LoggingDelivery struct {
	logger *zap.Logger
}

func NewLoggingDelivery(logger *zap.Logger) *LoggingDelivery {
	return &LoggingDelivery{logger: logger}
}

func (l LoggingDelivery) Send(ctx context.Context, req interfaces.SendRequest) error {
	if req.Message == nil {
		return nil
	}
	l.logger.Info("message received",
		zap.String("content_topic", req.Message.ContentTopic),
		zap.String("message", string(req.Message.Message)),
	)
	return nil
}
