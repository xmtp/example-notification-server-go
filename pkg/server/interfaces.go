package server

import (
	"time"

	v1 "github.com/xmtp/proto/go/message_api/v1"
)

type DeliveryMechanismKind int64

const (
	APNS DeliveryMechanismKind = 0
	FCM  DeliveryMechanismKind = 1
)

type DeliveryMechanism struct {
	Kind  DeliveryMechanismKind
	Value string
}

type RegisterResponse struct {
	InstallationId string
	ValidUntil     time.Time
}

type LookupInstallationsRequest struct {
	InstallationIds []string
}

type Installation struct {
	InstallationId    string
	DeliveryMechanism DeliveryMechanism
}

type Subscription struct {
	ID             int64
	CreatedAt      time.Time
	InstallationId string
	Topic          string
	IsActive       bool
}

// Pluggable Installation Service interface
type InstallationService interface {
	Register(installationId string, mechanism DeliveryMechanism) (RegisterResponse, error)
	Delete(installationId string) error
	GetInstallations(installationIds []string) ([]Installation, error)
}

// This interface is not expected to be pluggable
type SubscriptionService interface {
	Subscribe(installationId string, topics []string) error
	Unsubscribe(installationId string, topics []string) error
	GetSubscriptions(topic string) ([]Subscription, error)
}

// Pluggable interface for sending push notifications
type DeliveryService interface {
	Send(installations []Installation, message v1.Envelope)
}
