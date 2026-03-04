package xmtp

import (
	"fmt"

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
		return fmt.Errorf("could not get subscriptions: %w", err)
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
		return fmt.Errorf("could not get installations: %w", err)
	}

	if len(installations) == 0 {
		l.logger.Debug("No matching installations found for topic", zap.String("topic", env.ContentTopic))
		return nil
	}

	sendRequests := buildSendRequests(env, installations, subs)
	for _, request := range sendRequests {
		if !shouldDeliver(request.MessageContext, request.Subscription) {
			l.logger.Info("skipping delivery of request",
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

	info, ok, err := parseV4Envelope(env)
	if err != nil {
		return fmt.Errorf("could not parse envelope: %w", err)
	}

	if !ok {
		// What we have is not a group message or a welcome message.
		return nil
	}

	l.logger.Info("processing envelope", zap.String("topic", info.Topic))

	// TODO: thirtyDayPeriodsFromEpoch
	subs, err := l.subscriptions.GetSubscriptions(l.ctx, info.Topic, 0)
	if err != nil {
		return fmt.Errorf("could not get subscriptions: %w", err)
	}

	if len(subs) == 0 {
		l.logger.Debug("no matching subscriptions found for topic", zap.String("topic", info.Topic))
		return nil
	}

	installationIds := make([]string, len(subs))
	for i, sub := range subs {
		installationIds[i] = sub.InstallationId
	}

	installations, err := l.installations.GetInstallations(l.ctx, installationIds)
	if err != nil {
		return fmt.Errorf("could not get installations: %w", err)
	}

	if len(installations) == 0 {
		l.logger.Debug("no matching installations found for topic", zap.String("topic", info.Topic))
		return nil
	}

	requests := buildSendRequestV4(env, *info, installations, subs)
	for _, req := range requests {
		if !shouldDeliver(*info, req.Subscription) && false {
			l.logger.Debug("skipping delivery",
				zap.Any("message_context", *info),
				zap.Bool("subscription_has_hmac_key", req.Subscription.HmacKey != nil),
			)
			continue
		}

		l.logger.Info("delivering notification",
			zap.Any("send_request", req),
		)

		// err = l.deliver(req)
		// if err != nil {
		// 	l.logger.Error("error delivering request", zap.Error(err), zap.String("content_topic", req.MessageContext.Topic))
		// }
	}

	return nil
}
