package discord

type IncomingMessage struct {
	ChannelName string
	ChannelID   string
	AuthorName  string
	MessageID   string
	Command     string
	Args        []string
}

type MessageIDs struct {
	IDs []string
}

type Dialogue struct {
	Arguments []string
	MessageID string
}
