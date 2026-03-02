package xmtp

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	v1 "github.com/xmtp/example-notification-server-go/pkg/proto/message_api/v1"
)

type SubscriberV3 interface {
	// Subscribe to a stream of all messages
	SubscribeAll(context.Context, *v1.SubscribeAllRequest, ...grpc.CallOption) (
		v1.MessageApi_SubscribeAllClient,
		error,
	)
}

func NewV3Client(conn grpc.ClientConnInterface) SubscriberV3 {
	return v1.NewMessageApiClient(conn)
}

type v3Stream struct {
	inner v1.MessageApi_SubscribeAllClient
}

func newV3Stream(s v1.MessageApi_SubscribeAllClient) SubscriberStream {
	return &v3Stream{
		inner: s,
	}
}

func (s *v3Stream) Receive() (*EnvelopesWrapped, error) {

	env, err := s.inner.Recv()
	if err != nil {
		return nil, fmt.Errorf("could not receive envelopes: %w", err)
	}

	return &EnvelopesWrapped{
		V3: []*v1.Envelope{env},
	}, nil
}
