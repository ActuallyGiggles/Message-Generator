package discord

type IncomingMessage struct {
	ChannelName string
	ChannelID   string
	AuthorName  string
	MessageID   string
	Command     string
	Args        []string
}
