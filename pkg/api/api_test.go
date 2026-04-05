package api

import (
	"bytes"
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

	topicpkg "github.com/xmtp/xmtpd/pkg/topic"
)

const INSTALLATION_ID = "install1"
const testGroupTopic = "/xmtp/mls/1/g-24ce39d660600b3a98adff3075b6d1f4/proto"
const testWelcomeTopic = "/xmtp/mls/1/w-abcdef0123456789/proto"

type testContext struct {
	client            protoconnect.NotificationsClient
	ctx               context.Context
	httpClient        *http.Client
	installationsMock *mocks.Installations
	subscriptionsMock *mocks.Subscriptions
	apiServer         *ApiServer
}


func matchTopics(expected ...*topicpkg.Topic) interface{} {
	return mock.MatchedBy(func(actual []*topicpkg.Topic) bool {
		if len(actual) != len(expected) {
			return false
		}
		for i := range expected {
			if actual[i] == nil || expected[i] == nil {
				if actual[i] != expected[i] {
					return false
				}
				continue
			}
			if actual[i].Kind() != expected[i].Kind() || !bytes.Equal(actual[i].Bytes(), expected[i].Bytes()) {
				return false
			}
		}
		return true
	})
}

func matchSubscriptionInputs(expected ...interfaces.SubscriptionInput) interface{} {
	return mock.MatchedBy(func(actual []interfaces.SubscriptionInput) bool {
		if len(actual) != len(expected) {
			return false
		}
		for i := range expected {
			exp := expected[i]
			got := actual[i]
			if (got.Topic == nil) != (exp.Topic == nil) {
				return false
			}
			if got.Topic != nil {
				if got.Topic.Kind() != exp.Topic.Kind() || !bytes.Equal(got.Topic.Bytes(), exp.Topic.Bytes()) {
					return false
				}
			}
			if got.IsSilent != exp.IsSilent {
				return false
			}
			if len(got.HmacKeys) != len(exp.HmacKeys) {
				return false
			}
			for j := range exp.HmacKeys {
				if got.HmacKeys[j].ThirtyDayPeriodsSinceEpoch != exp.HmacKeys[j].ThirtyDayPeriodsSinceEpoch {
					return false
				}
				if !bytes.Equal(got.HmacKeys[j].Key, exp.HmacKeys[j].Key) {
					return false
				}
			}
		}
		return true
	})
}

func setupTest(t *testing.T) testContext {
	t.Helper()
	return setupTestWithListenerType(t, interfaces.ListenerTypeV3)
}

func setupTestWithListenerType(t *testing.T, listenerType interfaces.ListenerType) testContext {
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
	apiServer := NewApiServer(testutils.TestLogger(t), options.ApiOptions{Port: port}, installationsMock, subscriptionsMock, listenerType)
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
		interfaces.ListenerTypeV3,
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
		mock.MatchedBy(func(inst interfaces.Installation) bool {
			return inst.Id == INSTALLATION_ID &&
				inst.DeliveryMechanism.Kind == interfaces.APNS &&
				inst.DeliveryMechanism.Token == deviceToken &&
				inst.PayloadFormat == interfaces.PayloadFormatV3
		}),
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

	parsed := testutils.MustParseTopic(t, testGroupTopic)

	ctx.subscriptionsMock.On(
		"Subscribe",
		mock.Anything,
		INSTALLATION_ID,
		matchTopics(parsed),
	).Return(nil)

	_, err := ctx.client.Subscribe(
		ctx.ctx,
		connect.NewRequest(&proto.SubscribeRequest{
			InstallationId: INSTALLATION_ID,
			Topics:         []string{testGroupTopic},
		}),
	)

	require.NoError(t, err)
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

	parsed := testutils.MustParseTopic(t, testGroupTopic)

	ctx.subscriptionsMock.On(
		"Unsubscribe",
		mock.Anything,
		INSTALLATION_ID,
		matchTopics(parsed),
	).Return(nil)

	_, err := ctx.client.Unsubscribe(
		ctx.ctx,
		connect.NewRequest(&proto.UnsubscribeRequest{
			InstallationId: INSTALLATION_ID,
			Topics:         []string{testGroupTopic},
		}),
	)

	require.NoError(t, err)
}

func Test_Subscribe_BytesTopics(t *testing.T) {
	ctx := setupTest(t)

	parsed := testutils.MustParseTopic(t, testGroupTopic)
	ctx.subscriptionsMock.On("Subscribe", mock.Anything, INSTALLATION_ID, matchTopics(parsed)).Return(nil)
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

	parsed := testutils.MustParseTopic(t, testGroupTopic)
	ctx.subscriptionsMock.On("Subscribe", mock.Anything, INSTALLATION_ID, matchTopics(parsed)).Return(nil)
	_, err := ctx.client.Subscribe(ctx.ctx, connect.NewRequest(&proto.SubscribeRequest{
		InstallationId: INSTALLATION_ID,
		Topics:         []string{testGroupTopic},
		TopicsBytes:    [][]byte{parsed.Bytes()},
	}))
	require.NoError(t, err)
}

func Test_Subscribe_EmptyTopics(t *testing.T) {
	ctx := setupTest(t)

	ctx.subscriptionsMock.On("Subscribe", mock.Anything, INSTALLATION_ID, matchTopics()).Return(nil)
	_, err := ctx.client.Subscribe(ctx.ctx, connect.NewRequest(&proto.SubscribeRequest{
		InstallationId: INSTALLATION_ID,
	}))
	require.NoError(t, err)
}

func Test_Unsubscribe_BytesTopics(t *testing.T) {
	ctx := setupTest(t)

	parsed := testutils.MustParseTopic(t, testGroupTopic)
	ctx.subscriptionsMock.On("Unsubscribe", mock.Anything, INSTALLATION_ID, matchTopics(parsed)).Return(nil)
	_, err := ctx.client.Unsubscribe(ctx.ctx, connect.NewRequest(&proto.UnsubscribeRequest{
		InstallationId: INSTALLATION_ID,
		TopicsBytes:    [][]byte{parsed.Bytes()},
	}))
	require.NoError(t, err)
}

func Test_SubscribeWithMetadata_StringTopic(t *testing.T) {
	ctx := setupTest(t)

	ctx.subscriptionsMock.On(
		"SubscribeWithMetadata",
		mock.Anything,
		INSTALLATION_ID,
		matchSubscriptionInputs(interfaces.SubscriptionInput{
			Topic:    testutils.MustParseTopic(t, testGroupTopic),
			IsSilent: true,
		}),
	).Return(nil)
	_, err := ctx.client.SubscribeWithMetadata(ctx.ctx, connect.NewRequest(&proto.SubscribeWithMetadataRequest{
		InstallationId: INSTALLATION_ID,
		Subscriptions: []*proto.Subscription{{
			Topic:    testGroupTopic,
			IsSilent: true,
		}},
	}))
	require.NoError(t, err)
}

func Test_SubscribeWithMetadata_BytesTakesPrecedence(t *testing.T) {
	ctx := setupTest(t)

	parsed := testutils.MustParseTopic(t, testWelcomeTopic)
	ctx.subscriptionsMock.On(
		"SubscribeWithMetadata",
		mock.Anything,
		INSTALLATION_ID,
		matchSubscriptionInputs(interfaces.SubscriptionInput{
			Topic: parsed,
		}),
	).Return(nil)
	_, err := ctx.client.SubscribeWithMetadata(ctx.ctx, connect.NewRequest(&proto.SubscribeWithMetadataRequest{
		InstallationId: INSTALLATION_ID,
		Subscriptions: []*proto.Subscription{{
			Topic:      testGroupTopic,
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

func TestRegisterInstallation_WithPayloadFormatV4_OnV3Listener_ReturnsError(t *testing.T) {
	ctx := setupTestWithListenerType(t, interfaces.ListenerTypeV3)

	_, err := ctx.client.RegisterInstallation(
		ctx.ctx,
		connect.NewRequest(&proto.RegisterInstallationRequest{
			InstallationId: INSTALLATION_ID,
			DeliveryMechanism: &proto.DeliveryMechanism{
				DeliveryMechanismType: &proto.DeliveryMechanism_ApnsDeviceToken{ApnsDeviceToken: "token"},
			},
			PayloadFormat: proto.PayloadFormat_PAYLOAD_FORMAT_V4,
		}),
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid_argument")
}

func TestRegisterInstallation_WithPayloadFormatV4_OnV4Listener_Succeeds(t *testing.T) {
	ctx := setupTestWithListenerType(t, interfaces.ListenerTypeV4)

	validUntil := time.Now()
	ctx.installationsMock.On(
		"Register",
		mock.Anything,
		mock.MatchedBy(func(inst interfaces.Installation) bool {
			return inst.Id == INSTALLATION_ID &&
				inst.DeliveryMechanism.Kind == interfaces.APNS &&
				inst.DeliveryMechanism.Token == "token" &&
				inst.PayloadFormat == interfaces.PayloadFormatV4
		}),
	).Return(&interfaces.RegisterResponse{
		InstallationId: INSTALLATION_ID,
		ValidUntil:     validUntil,
	}, nil)

	result, err := ctx.client.RegisterInstallation(
		ctx.ctx,
		connect.NewRequest(&proto.RegisterInstallationRequest{
			InstallationId: INSTALLATION_ID,
			DeliveryMechanism: &proto.DeliveryMechanism{
				DeliveryMechanismType: &proto.DeliveryMechanism_ApnsDeviceToken{ApnsDeviceToken: "token"},
			},
			PayloadFormat: proto.PayloadFormat_PAYLOAD_FORMAT_V4,
		}),
	)

	require.NoError(t, err)
	require.Equal(t, INSTALLATION_ID, result.Msg.InstallationId)
}

func TestRegisterInstallation_WithUnspecified_DefaultsToV3(t *testing.T) {
	ctx := setupTest(t)


	validUntil := time.Now()
	ctx.installationsMock.On(
		"Register",
		mock.Anything,
		mock.MatchedBy(func(inst interfaces.Installation) bool {
			return inst.Id == INSTALLATION_ID &&
				inst.DeliveryMechanism.Kind == interfaces.APNS &&
				inst.DeliveryMechanism.Token == "token" &&
				inst.PayloadFormat == interfaces.PayloadFormatV3
		}),
	).Return(&interfaces.RegisterResponse{
		InstallationId: INSTALLATION_ID,
		ValidUntil:     validUntil,
	}, nil)

	result, err := ctx.client.RegisterInstallation(
		ctx.ctx,
		connect.NewRequest(&proto.RegisterInstallationRequest{
			InstallationId: INSTALLATION_ID,
			DeliveryMechanism: &proto.DeliveryMechanism{
				DeliveryMechanismType: &proto.DeliveryMechanism_ApnsDeviceToken{ApnsDeviceToken: "token"},
			},
			PayloadFormat: proto.PayloadFormat_PAYLOAD_FORMAT_UNSPECIFIED,
		}),
	)

	require.NoError(t, err)
	require.Equal(t, INSTALLATION_ID, result.Msg.InstallationId)
}
