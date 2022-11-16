package xmtp

import (
	"context"
	"fmt"

	v1 "github.com/xmtp/proto/go/message_api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func newConn(ctx context.Context, apiAddress string) (*grpc.ClientConn, error) {
	dialAddr := fmt.Sprintf(apiAddress)
	return grpc.DialContext(
		ctx,
		dialAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

func NewClient(ctx context.Context, apiAddress string) (v1.MessageApiClient, error) {
	conn, err := newConn(ctx, apiAddress)
	if err != nil {
		return nil, err
	}

	return v1.NewMessageApiClient(conn), nil
}
