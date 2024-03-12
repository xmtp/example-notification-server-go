package db

import (
	"time"

	"github.com/uptrace/bun"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
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

	ID             int64                            `bun:",pk,autoincrement"`
	CreatedAt      time.Time                        `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt      time.Time                        `bun:"updated_at,notnull,default:current_timestamp"`
	InstallationId string                           `bun:"installation_id,notnull,unique:group"`
	Installation   Installation                     `bun:"rel:belongs-to,join:installation_id=id"`
	Kind           interfaces.DeliveryMechanismKind `bun:"kind,notnull,unique:group"`
	Token          string                           `bun:"token,notnull,unique:group"`
}

type Subscription struct {
	bun.BaseModel `bun:"table:subscriptions"`

	Id             int64                   `bun:",pk,autoincrement"`
	CreatedAt      time.Time               `bun:"created_at,notnull,default:current_timestamp"`
	InstallationId string                  `bun:"installation_id,notnull"`
	Topic          string                  `bun:"topic,notnull"`
	IsActive       bool                    `bun:"is_active,notnull"`
	IsSilent       bool                    `bun:"is_silent,notnull"`
	HmacKeys       []*SubscriptionHmacKeys `bun:"rel:has-many,join:id=subscription_id"`
}

type SubscriptionHmacKeys struct {
	bun.BaseModel              `bun:"table:subscription_hmac_keys"`
	SubscriptionId             int64     `bun:"subscription_id,notnull,pk"`
	ThirtyDayPeriodsSinceEpoch int32     `bun:"thirty_day_periods_since_epoch,notnull,pk"`
	Key                        []byte    `bun:"key,notnull,type:bytea"`
	CreatedAt                  time.Time `bun:"created_at,notnull"`
	UpdatedAt                  time.Time `bun:"updated_at,notnull,default:current_timestamp"`
}
