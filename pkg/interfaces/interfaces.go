package interfaces

import (
	"context"
	"time"

	"github.com/xmtp/example-notification-server-go/pkg/db"
	v1 "github.com/xmtp/proto/go/message_api/v1"
)

type DeliveryMechanism struct {
	Kind      db.DeliveryMechanismKind
	Token     string
	UpdatedAt time.Time
}

type RegisterResponse struct {
	InstallationId string
	ValidUntil     time.Time
}

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
