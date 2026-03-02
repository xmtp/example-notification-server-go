package xmtp

import (
	"context"
	"crypto/tls"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func newConn(apiAddress string, useTls bool, clientVersion, appVersion string) (*grpc.ClientConn, error) {
	return grpc.NewClient(
		apiAddress,
		grpc.WithTransportCredentials(getCredentials(useTls)),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 5 * time.Second,
		}),
		grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			ctx = metadata.AppendToOutgoingContext(ctx, clientVersionMetadataKey, clientVersion)
			ctx = metadata.AppendToOutgoingContext(ctx, appVersionMetadataKey, appVersion)
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	)
}

func getCredentials(useTls bool) credentials.TransportCredentials {
	if useTls {
		return credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: false,
		})
	}
	return insecure.NewCredentials()
}
