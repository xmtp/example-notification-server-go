package xmtp

import (
	"errors"

	"go.uber.org/zap"

	v1 "github.com/xmtp/example-notification-server-go/pkg/proto/message_api/v1"
	"github.com/xmtp/example-notification-server-go/pkg/proto/xmtpv4/envelopes"
)

func (l *Listener) processV3Envelope(env *v1.Envelope) error {

	if !isV3Topic(env.ContentTopic) {
		l.logger.Debug("ignoring message", zap.String("topic", env.ContentTopic))
		return nil
	}

	subs, err := l.subscriptions.GetSubscriptions(l.ctx, env.ContentTopic, getThirtyDayPeriodsFromEpoch(env))
	if err != nil {
		return err
	}

	if len(subs) == 0 {
		return nil
	}

	installationIds := make([]string, len(subs))
	for i, sub := range subs {
		installationIds[i] = sub.InstallationId
	}

	installations, err := l.installations.GetInstallations(l.ctx, installationIds)
	if err != nil {
		return err
	}

	if len(installations) == 0 {
		l.logger.Info("No matching installations found for topic", zap.String("topic", env.ContentTopic))
		return nil
	}

	sendRequests := buildSendRequests(env, installations, subs)
	for _, request := range sendRequests {
		if !l.shouldDeliver(request.MessageContext, request.Subscription) {
			l.logger.Info("Skipping delivery of request",
				zap.Any("message_context", request.MessageContext),
				zap.Bool("subscription_has_hmac_key", request.Subscription.HmacKey != nil),
			)
			continue
		}
		if err = l.deliver(request); err != nil {
			l.logger.Error("error delivering request", zap.Error(err), zap.String("content_topic", env.ContentTopic))
		}
	}
	return err
}

func (l *Listener) processV4Envelope(env *envelopes.OriginatorEnvelope) error {

	return errors.New("TBD: not implemented")
}
