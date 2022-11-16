package interfaces

import (
	"context"
	"time"

	v1 "github.com/xmtp/proto/go/message_api/v1"
)

type DeliveryMechanismKind string

const (
	APNS DeliveryMechanismKind = "apns"
	FCM  DeliveryMechanismKind = "fcm"
)

type DeliveryMechanism struct {
	Kind      DeliveryMechanismKind
	Token     string
	UpdatedAt time.Time
}

type RegisterResponse struct {
	InstallationId string
	ValidUntil     time.Time
}

/**
An installation represents an app installed on a device. If the app is reinstalled, or installed onto
a new device it is expected to generate a fresh installation_id.
*/
type Installation struct {
	Id                string
	DeliveryMechanism DeliveryMechanism
}

type Subscription struct {
	ID             int64
	CreatedAt      time.Time
	InstallationId string
	Topic          string
	IsActive       bool
}

type SendRequest struct {
	IdempotencyKey string
	Installations  []Installation
	Message        v1.Envelope
}

type Installations interface {
	Register(ctx context.Context, installation Installation) (*RegisterResponse, error)
	Delete(ctx context.Context, installationId string) error
	GetInstallations(ctx context.Context, installationIds []string) ([]Installation, error)
}

// Pluggable Installation Service interface

// This interface is not expected to be pluggable
type Subscriptions interface {
	Subscribe(ctx context.Context, installationId string, topics []string) error
	Unsubscribe(ctx context.Context, installationId string, topics []string) error
	GetSubscriptions(ctx context.Context, topic string) ([]Subscription, error)
}

// Pluggable interface for sending push notifications
type Delivery interface {
	Send(ctx context.Context, req SendRequest)
}
