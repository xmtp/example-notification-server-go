package db

import (
	"time"

	"github.com/uptrace/bun"
)

type Installation struct {
	bun.BaseModel `bun:"table:installations"`

	ID        string     `bun:"id,pk"`
	CreatedAt time.Time  `bun:"created_at,notnull"`
	DeletedAt *time.Time `bun:"deleted_at"`
}

type DeviceDeliveryMechanism struct {
	bun.BaseModel `bun:"table:device_delivery_mechanisms"`

	ID             int64        `bun:",pk,autoincrement"`
	CreatedAt      time.Time    `bun:"created_at,notnull"`
	InstallationId string       `bun:"installation_id,notnull"`
	Installation   Installation `bun:"rel:belongs-to,join:installation_id=id"`
	Kind           string       `bun:"kind,notnull"`
	Token          string       `bun:"token,notnull"`
}

type Subscription struct {
	bun.BaseModel `bun:"table:subscriptions"`

	ID             int64     `bun:",pk,autoincrement"`
	CreatedAt      time.Time `bun:"created_at,notnull"`
	InstallationId string    `bun:"installation_id,notnull"`
	Topic          string    `bun:"topic,notnull"`
	IsActive       bool      `bun:"is_active,notnull"`
}