package xmtp

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/xmtp/example-notification-server-go/pkg/proto/xmtpv4/message_api"
)

type SubscriberV4 interface {
	SubscribeAllEnvelopes(context.Context, *message_api.SubscribeAllEnvelopesRequest, ...grpc.CallOption) (
		message_api.ReplicationApi_SubscribeAllEnvelopesClient,
		error,
	)
}

func NewV4Client(conn grpc.ClientConnInterface) SubscriberV4 {
	return message_api.NewReplicationApiClient(conn)
}

type v4Stream struct {
	inner message_api.ReplicationApi_SubscribeAllEnvelopesClient
}

func newV4Stream(s message_api.ReplicationApi_SubscribeAllEnvelopesClient) SubscriberStream {
	return &v4Stream{
		inner: s,
	}
}

func (s *v4Stream) Receive() (*EnvelopesWrapped, error) {

	res, err := s.inner.Recv()
	if err != nil {
		return nil, fmt.Errorf("could not receive envelopes: %w", err)
	}

	return &EnvelopesWrapped{
		V4: res.Envelopes,
	}, nil
}
