package xmtp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xmtp/example-notification-server-go/mocks"
	"github.com/xmtp/example-notification-server-go/pkg/installations"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/logging"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	"github.com/xmtp/example-notification-server-go/pkg/subscriptions"
	"github.com/xmtp/example-notification-server-go/test"
)

func TestV4Listener_NewAndStop(t *testing.T) {
	logger := logging.CreateLogger("console", "info")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db := test.CreateTestDb(t)
	instSvc := installations.NewInstallationsService(logger, db)
	subsSvc := subscriptions.NewSubscriptionsService(logger, db)
	mockDelivery := mocks.NewDelivery(t)

	l, err := NewV4Listener(ctx, logger, options.XmtpOptions{
		ListenerEnabled: true,
		GrpcAddress:     "localhost:25556",
		NumWorkers:      5,
	}, instSvc, subsSvc, []interfaces.Delivery{mockDelivery}, "test", "test")
	require.NoError(t, err)
	require.NotNil(t, l)
	l.Stop()
}
