package topics

type MessageType string

const (
	Test           MessageType = "test"
	Private        MessageType = "private"
	Contact        MessageType = "contact"
	V1Intro        MessageType = "v1-intro"
	V2Invite       MessageType = "v2-invite"
	V1Conversation MessageType = "v1-conversation"
	V2Conversation MessageType = "v2-conversation"
	V3Welcome      MessageType = "v3-welcome"
	V3Conversation MessageType = "v3-conversation"
	Unknown        MessageType = "unknown"
)
