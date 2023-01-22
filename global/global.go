package global

import (
	"os"
	"regexp"
	"sync"

	"github.com/joho/godotenv"
)

var (
	// Basic bot info
	Prefix  string
	BotName string

	// Twitch API information
	TwitchOAuth    string
	TwitchClientID string

	// Youtube API information
	YoutubeAPIKey       string
	YoutubeClientID     string
	YoutubeClientSecret string

	// Emote list
	GlobalEmotes            []Emote
	TwitchChannelEmotes     []Emote
	ThirdPartyChannelEmotes []ThirdPartyEmotes
	EmotesMx                sync.Mutex

	// Discord variables
	DiscordToken                   string
	DiscordGuildID                 string
	DiscordOwnerID                 string
	DiscordBotID                   string
	DiscordModChannelID            string
	DiscordTweetEmote              string
	DiscordAllChannelID            string
	DiscordReplyChannelID          string
	DiscordParticipationChannelID  string
	DiscordErrorTrackingChannelID  string
	DiscordWebsiteResultsChannelID string

	// Twitter variables
	TwitterAPIKey            string
	TwitterAPISecret         string
	TwitterAccessToken       string
	TwitterAccessTokenSecret string
	TwitterClientID          string
	TwitterClientSecret      string
	TwitterBearerToken       string

	Directives []Directive

	BannedUsers []string
	RegexList   []string
	Regex       *regexp.Regexp
)

func Start() {
	// Load ENV Variables
	godotenv.Load(".env")

	Prefix = os.Getenv("PREFIX")
	BotName = os.Getenv("BOT_NAME")

	// Twitch
	TwitchOAuth = os.Getenv("TWITCH_OAUTH")
	TwitchClientID = os.Getenv("TWITCH_CLIENT_ID")

	// Discord
	DiscordToken = os.Getenv("DISCORD_TOKEN")
	DiscordGuildID = os.Getenv("DISCORD_GUILD_ID")
	DiscordOwnerID = os.Getenv("DISCORD_OWNER_ID")
	DiscordBotID = os.Getenv("DISCORD_BOT_ID")
	DiscordModChannelID = os.Getenv("DISCORD_MOD_CHANNEL_ID")
	DiscordTweetEmote = os.Getenv("DISCORD_TWEET_EMOTE")

	// Twitter
	TwitterAccessToken = os.Getenv("TWITTER_ACCESS_TOKEN")
	TwitterAccessTokenSecret = os.Getenv("TWITTER_ACCESS_TOKEN_SECRET")

	LoadChannels()
	LoadRegex()
	LoadBannedUsers()
}
