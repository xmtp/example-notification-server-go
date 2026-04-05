package xmtp

import (
	"context"
	"crypto/tls"
	"time"

	v1 "github.com/xmtp/xmtpd/pkg/proto/message_api/v1"
	notificationApi "github.com/xmtp/xmtpd/pkg/proto/xmtpv4/message_api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	clientVersionMetadataKey = "x-client-version"
	appVersionMetadataKey    = "x-app-version"
)

func newConn(apiAddress string, useTls bool, clientVersion, appVersion string) (*grpc.ClientConn, error) {
	return grpc.NewClient(
		apiAddress,
		grpc.WithTransportCredentials(getCredentials(useTls)),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 5 * time.Second,
		}),
		grpc.WithUnaryInterceptor(metadataUnaryInterceptor(clientVersion, appVersion)),
		grpc.WithStreamInterceptor(metadataStreamInterceptor(clientVersion, appVersion)),
	)
}

func appendVersionMetadata(ctx context.Context, clientVersion, appVersion string) context.Context {
	ctx = metadata.AppendToOutgoingContext(ctx, clientVersionMetadataKey, clientVersion)
	ctx = metadata.AppendToOutgoingContext(ctx, appVersionMetadataKey, appVersion)
	return ctx
}

func metadataUnaryInterceptor(clientVersion, appVersion string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		return invoker(appendVersionMetadata(ctx, clientVersion, appVersion), method, req, reply, cc, opts...)
	}
}

func metadataStreamInterceptor(clientVersion, appVersion string) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		return streamer(appendVersionMetadata(ctx, clientVersion, appVersion), desc, cc, method, opts...)
	}
}

func getCredentials(useTls bool) credentials.TransportCredentials {
	if useTls {
		return credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: false,
		})
	}
	return insecure.NewCredentials()
}

func NewClient(ctx context.Context, apiAddress string, useTls bool, clientVersion, appVersion string) (v1.MessageApiClient, error) {
	conn, err := newConn(apiAddress, useTls, clientVersion, appVersion)
	if err != nil {
		return nil, err
	}

	return v1.NewMessageApiClient(conn), nil
}

func NewV4Client(ctx context.Context, apiAddress string, useTls bool, clientVersion, appVersion string) (notificationApi.NotificationApiClient, *grpc.ClientConn, error) {
	conn, err := newConn(apiAddress, useTls, clientVersion, appVersion)
	if err != nil {
		return nil, nil, err
	}

	return notificationApi.NewNotificationApiClient(conn), conn, nil
}
