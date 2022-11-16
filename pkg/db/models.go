package db

import (
	"time"

	"github.com/uptrace/bun"
)

type DeliveryMechanismKind string

const (
	APNS DeliveryMechanismKind = "apns"
	FCM  DeliveryMechanismKind = "fcm"
)

type Installation struct {
	bun.BaseModel `bun:"table:installations,select:installations"`

	Id                 string                     `bun:",pk"`
	CreatedAt          time.Time                  `bun:"created_at,notnull,default:current_timestamp"`
	DeletedAt          *time.Time                 `bun:"deleted_at"`
	DeliveryMechanisms []*DeviceDeliveryMechanism `bun:"rel:has-many,join:id=installation_id"`
}

type DeviceDeliveryMechanism struct {
	bun.BaseModel `bun:"table:device_delivery_mechanisms"`

	CreatedAt      time.Time             `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt      time.Time             `bun:"updated_at,notnull,default:current_timestamp"`
	InstallationId string                `bun:"installation_id,notnull,pk"`
	Installation   Installation          `bun:"rel:belongs-to,join:installation_id=id"`
	Kind           DeliveryMechanismKind `bun:"kind,notnull,pk"`
	Token          string                `bun:"token,notnull,pk"`
}

type Subscription struct {
	bun.BaseModel `bun:"table:subscriptions"`

	ID             int64     `bun:",pk,autoincrement"`
	CreatedAt      time.Time `bun:"created_at,notnull,default:current_timestamp"`
	InstallationId string    `bun:"installation_id,notnull"`
	Topic          string    `bun:"topic,notnull"`
	IsActive       bool      `bun:"is_active,notnull"`
}
