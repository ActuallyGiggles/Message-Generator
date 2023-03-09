package global

type DiscordChannelInfo struct {
	ChannelName string
	ChannelID   string
}

type Directive struct {
	Platform         string
	ChannelName      string
	ChannelID        string
	DiscordChannelID string
	Settings         DirectiveSettings
}

type DirectiveSettings struct {
	IsCollectingMessages bool
	Reply                PostConditions
	Participation        PostConditions
	WhichChannelsToUse   string
	CustomChannelsToUse  []string
}

type PostConditions struct {
	IsEnabled            bool
	IsAllowedWhenOnline  bool
	OnlineTimeToWait     int
	IsAllowedWhenOffline bool
	OfflineTimeToWait    int
}

type Resource struct {
	DiscordChannelName string
	DiscordChannelID   string
	DisplayMessageID   string
	Content            string
}

type Emote struct {
	Name string
	Url  string
}

type ThirdPartyEmotes struct {
	Name   string
	Emotes []Emote
}
