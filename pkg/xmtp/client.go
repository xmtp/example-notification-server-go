package xmtp

import (
	"context"
	"fmt"

	v1 "github.com/xmtp/example-notification-server-go/pkg/proto/message_api/v1"
	"github.com/xmtp/example-notification-server-go/pkg/proto/xmtpv4/envelopes"
	"github.com/xmtp/example-notification-server-go/pkg/proto/xmtpv4/message_api"
	"google.golang.org/grpc"
)

const (
	clientVersionMetadataKey = "x-client-version"
	appVersionMetadataKey    = "x-app-version"
)

type SubscriberClient interface {
	Subscribe(ctx context.Context, cursor map[uint32]uint64) (SubscriberStream, error)
}

type SubscriberStream interface {
	Receive() (*EnvelopesWrapped, error)
}

type clientConfig struct {
	v3Client bool
}

type clientOption func(*clientConfig)

func UseV3Client(b bool) clientOption {
	return func(cfg *clientConfig) {
		cfg.v3Client = b
	}
}

type clientWrapper struct {
	useV4 bool

	v3sub SubscriberV3
	v4sub SubscriberV4
}

func newSubscriberClient(conn grpc.ClientConnInterface, opts ...clientOption) SubscriberClient {

	var cfg clientConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	// Use the legacy, v3 cliient
	if cfg.v3Client {

		return &clientWrapper{
			v3sub: NewV3Client(conn),
		}
	}

	return &clientWrapper{
		useV4: true,
		v4sub: NewV4Client(conn),
	}
}

// TODO: Implement a refresh

// Subscribe opens an envelope stream. Cursor is ignored for v3 streams.
func (c *clientWrapper) Subscribe(ctx context.Context, cursor map[uint32]uint64) (SubscriberStream, error) {

	if c.useV4 {

		req := &message_api.SubscribeAllEnvelopesRequest{
			LastSeen: &envelopes.Cursor{
				NodeIdToSequenceId: cursor,
			},
		}

		stream, err := c.v4sub.SubscribeAllEnvelopes(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("could not subscribe to envelopes: %w", err)
		}

		return newV4Stream(stream), nil
	}

	// Open the legacy, v3 stream.
	stream, err := c.v3sub.SubscribeAll(ctx, &v1.SubscribeAllRequest{})
	if err != nil {
		return nil, fmt.Errorf("could not subscribe to envelopes: %w", err)
	}

	return newV3Stream(stream), nil
}

type EnvelopesWrapped struct {
	V4 []*envelopes.OriginatorEnvelope
	V3 []*v1.Envelope
}
