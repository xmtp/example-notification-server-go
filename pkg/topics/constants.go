package topics

type MessageType string

const (
	Test           MessageType = "test"
	V3Welcome      MessageType = "v3-welcome"
	V3Conversation MessageType = "v3-conversation"
	Unknown        MessageType = "unknown"
)
