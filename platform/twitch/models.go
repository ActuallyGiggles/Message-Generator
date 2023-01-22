package twitch

type Broadcaster[T any] struct {
	Data []T
}

type Data struct {
	ID              string `json:"id"`
	Login           string `json:"login"`
	DisplayName     string `json:"display_name"`
	Type            string `json:"type"`
	BroadcasterType string `json:"broadcaster_type"`
	Description     string `json:"description"`
	ProfileImageUrl string `json:"profile_image_url"`
	OfflineImageUrl string `json:"offline_image_url"`
	ViewCount       int    `json:"view_count"`
	Email           string `json:"email"`
	CreatedAt       string `json:"created_at"`
}

type TwitchEmoteAPIResponse[T any] struct {
	Data     []T    `json:"data"`
	Template string `json:"template"`
}

type TwitchGlobalEmote struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Images    map[string]string `json:"images"`
	Format    []string          `json:"format"`
	Scale     []string          `json:"scale"`
	ThemeMode []string          `json:"theme_mode"`
}

type TwitchChannelEmote struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Images     map[string]string `json:"images"`
	Tier       string            `json:"tier"`
	EmoteType  string            `json:"emote_type"`
	EmoteSetID string            `json:"emote_set_id"`
	Format     []string          `json:"format"`
	Scale      []string          `json:"scale"`
	ThemeMode  []string          `json:"theme_mode"`
}

type SevenTVEmote struct {
	Name string     `json:"name"`
	Urls [][]string `json:"urls"`
}

type BttvChannelEmotes[T any] struct {
	ChannelEmotes []T `json:"channelEmotes"`
	SharedEmotes  []T `json:"sharedEmotes"`
}

type BttvEmote struct {
	Name string `json:"code"`
	ID   string `json:"id"`
}

type FfzSets struct {
	Sets map[string]FfzSet `json:"sets"`
}

type FfzSet struct {
	Emoticons []FfzEmotes `json:"emoticons"`
}

type FfzEmotes struct {
	Name string            `json:"name"`
	Urls map[string]string `json:"urls"`
}

type StreamStatusData struct {
	Data []StreamStatusActual `json:"data"`
}

type StreamStatusActual struct {
	Name string `json:"user_login"`
	Type string `json:"type"`
}
