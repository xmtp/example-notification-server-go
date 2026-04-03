package api

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/xmtp/example-notification-server-go/mocks"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/logging"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	proto "github.com/xmtp/example-notification-server-go/pkg/proto/notifications/v1"
	protoconnect "github.com/xmtp/example-notification-server-go/pkg/proto/notifications/v1/notificationsv1connect"
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
	apiServer := NewApiServer(logging.CreateLogger("console", "info"), options.ApiOptions{Port: 8080}, installationsMock, subscriptionsMock)
	apiServer.Start()
	time.Sleep(50 * time.Millisecond)

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

	tests := []struct {
		name             string
		topics           []string
		expectedTopicIDs []string
	}{
		{
			name:             "plain topic",
			topics:           []string{"topicName"},
			expectedTopicIDs: []string{"topicName"},
		},
		{
			name:             "old group message topic",
			topics:           []string{"/xmtp/mls/1/g-topic1"},
			expectedTopicIDs: []string{"topic1"},
		},
		{
			name:             "old welcome message topic",
			topics:           []string{"/xmtp/mls/1/w-topic3"},
			expectedTopicIDs: []string{"topic3"},
		},
		{
			name: "group of old topics",
			topics: []string{
				"/xmtp/mls/1/g-topic1",
				"/xmtp/mls/1/g-topic2",
				"/xmtp/mls/1/w-topic3",
				"/xmtp/mls/1/g-topic4/proto",
			},
			expectedTopicIDs: []string{
				"topic1",
				"topic2",
				"topic3",
				"topic4/proto",
			},
		},
		{
			name: "group of new topics",
			topics: []string{
				"topic1",
				"topic2",
				"topic3",
			},
			expectedTopicIDs: []string{
				"topic1",
				"topic2",
				"topic3",
			},
		},
		{
			name: "group of mixed topics",
			topics: []string{
				"/xmtp/mls/1/g-topic1",
				"/xmtp/mls/1/g-topic2",
				"/xmtp/mls/1/w-topic3",
				"topicA",
				"topicB",
			},
			expectedTopicIDs: []string{
				"topic1",
				"topic2",
				"topic3",
				"topicA",
				"topicB",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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
					Topics:         test.topics,
				}),
			)
			require.NoError(t, err)
			ctx.subscriptionsMock.AssertCalled(
				t,
				"Subscribe",
				mock.Anything,
				INSTALLATION_ID,
				test.expectedTopicIDs,
			)
		})
	}
}

func Test_SubscribeRejectsInvalidTopic(t *testing.T) {

	ctx := setupTest(t)
	defer ctx.cleanup()

	topics := []string{
		"/xmtp/mls/1/s-topic",
	}

	_, err := ctx.client.Subscribe(
		ctx.ctx,
		connect.NewRequest(&proto.SubscribeRequest{
			InstallationId: INSTALLATION_ID,
			Topics:         topics,
		}),
	)
	require.Error(t, err)
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
