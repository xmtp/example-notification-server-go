package xmtp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type noopClientStream struct {
	ctx context.Context
}

func (s noopClientStream) Header() (metadata.MD, error) { return nil, nil }
func (s noopClientStream) Trailer() metadata.MD         { return nil }
func (s noopClientStream) CloseSend() error             { return nil }
func (s noopClientStream) Context() context.Context     { return s.ctx }
func (s noopClientStream) SendMsg(any) error            { return nil }
func (s noopClientStream) RecvMsg(any) error            { return nil }

func TestMetadataUnaryInterceptor_AppendsVersionHeaders(t *testing.T) {
	interceptor := metadataUnaryInterceptor("client-test", "app-test")

	err := interceptor(
		t.Context(),
		"/xmtp.xmtpv4.message_api.NotificationApi/SubscribeAllEnvelopes",
		nil,
		nil,
		nil,
		func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			md, ok := metadata.FromOutgoingContext(ctx)
			require.True(t, ok)
			require.Equal(t, []string{"client-test"}, md.Get(clientVersionMetadataKey))
			require.Equal(t, []string{"app-test"}, md.Get(appVersionMetadataKey))
			return nil
		},
	)
	require.NoError(t, err)
}

func TestMetadataStreamInterceptor_AppendsVersionHeaders(t *testing.T) {
	interceptor := metadataStreamInterceptor("client-test", "app-test")

	stream, err := interceptor(
		t.Context(),
		&grpc.StreamDesc{ServerStreams: true},
		nil,
		"/xmtp.xmtpv4.message_api.NotificationApi/SubscribeAllEnvelopes",
		func(ctx context.Context, _ *grpc.StreamDesc, _ *grpc.ClientConn, _ string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
			md, ok := metadata.FromOutgoingContext(ctx)
			require.True(t, ok)
			require.Equal(t, []string{"client-test"}, md.Get(clientVersionMetadataKey))
			require.Equal(t, []string{"app-test"}, md.Get(appVersionMetadataKey))
			return noopClientStream{ctx: ctx}, nil
		},
	)
	require.NoError(t, err)
	require.NotNil(t, stream)
}
