package api

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/xmtp/example-notification-server-go/mocks"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/logging"
	"github.com/xmtp/example-notification-server-go/pkg/proto"
	"github.com/xmtp/example-notification-server-go/pkg/proto/protoconnect"
)

const INSTALLATION_ID = "install1"

func buildClient() protoconnect.NotificationsClient {
	return protoconnect.NewNotificationsClient(
		http.DefaultClient,
		"http://localhost:8080",
	)
}

type testContext struct {
	cleanup           func()
	client            protoconnect.NotificationsClient
	ctx               context.Context
	installationsMock *mocks.Installations
	subscriptionsMock *mocks.Subscriptions
	apiServer         *ApiServer
}

func setupTest(t *testing.T) testContext {
	ctx := context.Background()
	installationsMock := mocks.NewInstallations(t)
	subscriptionsMock := mocks.NewSubscriptions(t)
	apiServer := NewApiServer(logging.CreateLogger("console", "info"), 8080, installationsMock, subscriptionsMock)
	cleanup := func() {
		apiServer.Stop()
	}

	return testContext{
		cleanup:           cleanup,
		client:            buildClient(),
		ctx:               ctx,
		installationsMock: installationsMock,
		subscriptionsMock: subscriptionsMock,
		apiServer:         apiServer,
	}
}

func Test_RegisterInstallation(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.cleanup()

	deviceToken := "foo"
	validUntil := time.Now()

	ctx.installationsMock.On(
		"Register",
		mock.Anything,
		mock.Anything,
	).Return(&interfaces.RegisterResponse{
		InstallationId: INSTALLATION_ID,
		ValidUntil:     validUntil,
	}, nil)

	result, err := ctx.client.RegisterInstallation(
		ctx.ctx,
		connect.NewRequest(&proto.RegisterInstallationRequest{
			InstallationId: INSTALLATION_ID,
			DeliveryMechanism: &proto.DeliveryMechanism{
				DeliveryMechanismType: &proto.DeliveryMechanism_ApnsDeviceToken{ApnsDeviceToken: deviceToken},
			},
		}),
	)

	require.NoError(t, err)
	require.Equal(t, result.Msg.InstallationId, INSTALLATION_ID)
	require.Equal(t, result.Msg.ValidUntil, uint64(validUntil.UnixMilli()))
}

func Test_RegisterInstallationError(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.cleanup()

	ctx.installationsMock.On(
		"Register",
		mock.Anything,
		mock.Anything,
	).Return(nil, errors.New("err"))

	result, err := ctx.client.RegisterInstallation(
		ctx.ctx,
		connect.NewRequest(&proto.RegisterInstallationRequest{
			InstallationId: INSTALLATION_ID,
			DeliveryMechanism: &proto.DeliveryMechanism{
				DeliveryMechanismType: &proto.DeliveryMechanism_ApnsDeviceToken{ApnsDeviceToken: "foo"},
			},
		}),
	)

	require.Equal(t, err.Error(), "internal: err")
	require.Nil(t, result)
}

func Test_DeleteInstallation(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.cleanup()

	ctx.installationsMock.On("Delete", mock.Anything, mock.Anything).
		Return(nil)

	_, err := ctx.client.DeleteInstallation(
		ctx.ctx,
		connect.NewRequest(&proto.DeleteInstallationRequest{
			InstallationId: INSTALLATION_ID,
		}),
	)

	require.NoError(t, err)
	ctx.installationsMock.AssertCalled(
		t,
		"Delete",
		mock.Anything,
		INSTALLATION_ID,
	)
}

func Test_Subscribe(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.cleanup()
	topics := []string{"topic1"}

	ctx.subscriptionsMock.On(
		"Subscribe",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	_, err := ctx.client.Subscribe(
		ctx.ctx,
		connect.NewRequest(&proto.SubscribeRequest{
			InstallationId: INSTALLATION_ID,
			Topics:         topics,
		}),
	)

	require.NoError(t, err)
	ctx.subscriptionsMock.AssertCalled(
		t,
		"Subscribe",
		mock.Anything,
		INSTALLATION_ID,
		topics,
	)
}

func Test_SubscribeError(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.cleanup()

	ctx.subscriptionsMock.On(
		"Subscribe",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(errors.New("test"))

	_, err := ctx.client.Subscribe(
		ctx.ctx,
		connect.NewRequest(&proto.SubscribeRequest{
			InstallationId: INSTALLATION_ID,
			Topics:         []string{"topic1"},
		}),
	)

	require.Error(t, err)
	require.Equal(t, err.Error(), "internal: test")
}

func Test_Unsubscribe(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.cleanup()
	topics := []string{"topic1"}

	ctx.subscriptionsMock.On(
		"Unsubscribe",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	_, err := ctx.client.Unsubscribe(
		ctx.ctx,
		connect.NewRequest(&proto.UnsubscribeRequest{
			InstallationId: INSTALLATION_ID,
			Topics:         topics,
		}),
	)

	require.NoError(t, err)
	ctx.subscriptionsMock.AssertCalled(
		t,
		"Unsubscribe",
		mock.Anything,
		INSTALLATION_ID,
		topics,
	)
}
