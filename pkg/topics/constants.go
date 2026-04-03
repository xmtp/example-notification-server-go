package topics

type MessageType string

const (
	Test           MessageType = "test"
	V3Welcome      MessageType = "v3-welcome"
	V3Conversation MessageType = "v3-conversation"
	Unknown        MessageType = "unknown"
)

const (
	V3CommonPrefix         = "/xmtp/mls/1/"
	V3GroupPrefix          = "/xmtp/mls/1/g-"
	V3WelcomeMessagePrefix = "/xmtp/mls/1/w-"
)
