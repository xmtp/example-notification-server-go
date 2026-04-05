package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/xmtp/example-notification-server-go/mocks"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	"github.com/xmtp/example-notification-server-go/pkg/testutils"
	proto "github.com/xmtp/example-notification-server-go/pkg/proto/notifications/v1"
	protoconnect "github.com/xmtp/example-notification-server-go/pkg/proto/notifications/v1/notificationsv1connect"
)

const INSTALLATION_ID = "install1"

type testContext struct {
	client            protoconnect.NotificationsClient
	ctx               context.Context
	httpClient        *http.Client
	installationsMock *mocks.Installations
	subscriptionsMock *mocks.Subscriptions
	apiServer         *ApiServer
}

func setupTest(t *testing.T) testContext {
	t.Helper()
	ctx := t.Context()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	installationsMock := mocks.NewInstallations(t)
	subscriptionsMock := mocks.NewSubscriptions(t)
	httpClient := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
	apiServer := NewApiServer(testutils.TestLogger(t), options.ApiOptions{Port: port}, installationsMock, subscriptionsMock)
	require.NoError(t, apiServer.SetListener(listener))
	apiServer.Start()
	time.Sleep(50 * time.Millisecond)

	t.Cleanup(func() {
		httpClient.CloseIdleConnections()
		apiServer.Stop()
	})

	return testContext{
		client:            protoconnect.NewNotificationsClient(httpClient, fmt.Sprintf("http://127.0.0.1:%d", port)),
		ctx:               ctx,
		httpClient:        httpClient,
		installationsMock: installationsMock,
		subscriptionsMock: subscriptionsMock,
		apiServer:         apiServer,
	}
}

func Test_SetListenerAfterStartReturnsError(t *testing.T) {
	apiServer := NewApiServer(
		testutils.TestLogger(t),
		options.ApiOptions{Port: 18081},
		mocks.NewInstallations(t),
		mocks.NewSubscriptions(t),
	)
	apiServer.Start()
	defer apiServer.Stop()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, listener.Close())
	}()

	err = apiServer.SetListener(listener)
	require.EqualError(t, err, "api server already started")
}

func Test_RegisterInstallation(t *testing.T) {
	ctx := setupTest(t)

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
