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
	topicutil "github.com/xmtp/example-notification-server-go/pkg/topics"
	topicpkg "github.com/xmtp/xmtpd/pkg/topic"
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
	topics := []string{"/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto"}

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
		mock.Anything,
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
			Topics:         []string{"/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto"},
		}),
	)

	require.Error(t, err)
	require.Equal(t, err.Error(), "internal: test")
}

func Test_Unsubscribe(t *testing.T) {
	ctx := setupTest(t)
	topics := []string{"/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto"}

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
		mock.Anything,
	)
}

func Test_Subscribe_StringTopics(t *testing.T) {
	ctx := setupTest(t)

	topicStr := "/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto"
	ctx.subscriptionsMock.On("Subscribe", mock.Anything, INSTALLATION_ID, mock.Anything).Return(nil)
	_, err := ctx.client.Subscribe(ctx.ctx, connect.NewRequest(&proto.SubscribeRequest{
		InstallationId: INSTALLATION_ID,
		Topics:         []string{topicStr},
	}))
	require.NoError(t, err)
	ctx.subscriptionsMock.AssertCalled(t, "Subscribe", mock.Anything, INSTALLATION_ID, mock.Anything)
}

func Test_Subscribe_BytesTopics(t *testing.T) {
	ctx := setupTest(t)

	parsed, _ := topicutil.ParseV3Topic("/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto")
	ctx.subscriptionsMock.On("Subscribe", mock.Anything, INSTALLATION_ID, mock.Anything).Return(nil)
	_, err := ctx.client.Subscribe(ctx.ctx, connect.NewRequest(&proto.SubscribeRequest{
		InstallationId: INSTALLATION_ID,
		TopicsBytes:    [][]byte{parsed.Bytes()},
	}))
	require.NoError(t, err)
}

func Test_Subscribe_InvalidStringTopic(t *testing.T) {
	ctx := setupTest(t)

	_, err := ctx.client.Subscribe(ctx.ctx, connect.NewRequest(&proto.SubscribeRequest{
		InstallationId: INSTALLATION_ID,
		Topics:         []string{"invalid-topic"},
	}))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid")
}

func Test_Subscribe_InvalidBytesTopics(t *testing.T) {
	ctx := setupTest(t)

	_, err := ctx.client.Subscribe(ctx.ctx, connect.NewRequest(&proto.SubscribeRequest{
		InstallationId: INSTALLATION_ID,
		TopicsBytes:    [][]byte{{0xFF}}, // Too short, invalid kind
	}))
	require.Error(t, err)
}

func Test_Subscribe_MergedTopics(t *testing.T) {
	ctx := setupTest(t)

	parsed, _ := topicutil.ParseV3Topic("/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto")
	ctx.subscriptionsMock.On("Subscribe", mock.Anything, INSTALLATION_ID, mock.MatchedBy(func(topics []*topicpkg.Topic) bool {
		return len(topics) == 1 // deduplicated
	})).Return(nil)
	_, err := ctx.client.Subscribe(ctx.ctx, connect.NewRequest(&proto.SubscribeRequest{
		InstallationId: INSTALLATION_ID,
		Topics:         []string{"/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto"},
		TopicsBytes:    [][]byte{parsed.Bytes()},
	}))
	require.NoError(t, err)
}

func Test_Subscribe_EmptyTopics(t *testing.T) {
	ctx := setupTest(t)

	ctx.subscriptionsMock.On("Subscribe", mock.Anything, INSTALLATION_ID, mock.Anything).Return(nil)
	_, err := ctx.client.Subscribe(ctx.ctx, connect.NewRequest(&proto.SubscribeRequest{
		InstallationId: INSTALLATION_ID,
	}))
	require.NoError(t, err)
}

func Test_Unsubscribe_BytesTopics(t *testing.T) {
	ctx := setupTest(t)

	parsed, _ := topicutil.ParseV3Topic("/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto")
	ctx.subscriptionsMock.On("Unsubscribe", mock.Anything, INSTALLATION_ID, mock.MatchedBy(func(topics []*topicpkg.Topic) bool {
		return len(topics) == 1 &&
			topics[0].Kind() == topicpkg.TopicKindGroupMessagesV1 &&
			string(topics[0].Bytes()) == string(parsed.Bytes())
	})).Return(nil)
	_, err := ctx.client.Unsubscribe(ctx.ctx, connect.NewRequest(&proto.UnsubscribeRequest{
		InstallationId: INSTALLATION_ID,
		TopicsBytes:    [][]byte{parsed.Bytes()},
	}))
	require.NoError(t, err)
	ctx.subscriptionsMock.AssertCalled(t, "Unsubscribe", mock.Anything, INSTALLATION_ID, mock.Anything)
}

func Test_SubscribeWithMetadata_StringTopic(t *testing.T) {
	ctx := setupTest(t)

	ctx.subscriptionsMock.On("SubscribeWithMetadata", mock.Anything, INSTALLATION_ID, mock.Anything).Return(nil)
	_, err := ctx.client.SubscribeWithMetadata(ctx.ctx, connect.NewRequest(&proto.SubscribeWithMetadataRequest{
		InstallationId: INSTALLATION_ID,
		Subscriptions: []*proto.Subscription{{
			Topic:    "/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto",
			IsSilent: true,
		}},
	}))
	require.NoError(t, err)
	ctx.subscriptionsMock.AssertCalled(t, "SubscribeWithMetadata", mock.Anything, INSTALLATION_ID, mock.Anything)
}

func Test_SubscribeWithMetadata_BytesTakesPrecedence(t *testing.T) {
	ctx := setupTest(t)

	parsed, _ := topicutil.ParseV3Topic("/xmtp/mls/1/w-abcdef0123456789/proto")
	ctx.subscriptionsMock.On("SubscribeWithMetadata", mock.Anything, INSTALLATION_ID, mock.MatchedBy(func(inputs []interfaces.SubscriptionInput) bool {
		return len(inputs) == 1 && inputs[0].Topic.Kind() == topicpkg.TopicKindWelcomeMessagesV1
	})).Return(nil)
	_, err := ctx.client.SubscribeWithMetadata(ctx.ctx, connect.NewRequest(&proto.SubscribeWithMetadataRequest{
		InstallationId: INSTALLATION_ID,
		Subscriptions: []*proto.Subscription{{
			Topic:      "/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto",
			TopicBytes: parsed.Bytes(),
		}},
	}))
	require.NoError(t, err)
}

func Test_SubscribeWithMetadata_InvalidTopic(t *testing.T) {
	ctx := setupTest(t)

	_, err := ctx.client.SubscribeWithMetadata(ctx.ctx, connect.NewRequest(&proto.SubscribeWithMetadataRequest{
		InstallationId: INSTALLATION_ID,
		Subscriptions: []*proto.Subscription{{
			Topic: "not-a-valid-topic",
		}},
	}))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid")
}

func Test_SubscribeWithMetadata_EmptyTopic(t *testing.T) {
	ctx := setupTest(t)

	_, err := ctx.client.SubscribeWithMetadata(ctx.ctx, connect.NewRequest(&proto.SubscribeWithMetadataRequest{
		InstallationId: INSTALLATION_ID,
		Subscriptions:  []*proto.Subscription{{}},
	}))
	require.Error(t, err)
	require.Contains(t, err.Error(), "no topic")
}
