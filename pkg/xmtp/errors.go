package xmtp

import "errors"

var (
	ErrMissingClientEnvelope = errors.New("missing client envelope")
	ErrTopicMismatch         = errors.New("topic does not match payload")
	ErrFailedToMarshal       = errors.New("failed to marshal envelope")
	ErrUnknownWelcomeVersion = errors.New("unknown welcome version")
	ErrUnknownPayloadType    = errors.New("unknown payload type")
)
