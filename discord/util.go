package discord

import (
	"Twitch-Message-Generator/global"
	"Twitch-Message-Generator/platform/twitch"
	"Twitch-Message-Generator/stats"
	"Twitch-Message-Generator/twitter"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// wrapInCodeBlock returns a message wrapped in three back ticks for Discord formatting.
func WrapInCodeBlock(message string) string {
	return "```" + message + "```"
}

// memoryMonitor will send a message into error-tracking if the memory usage is exceeded a certain amount.
// func memoryMonitor() {
// 	for range time.Tick(10 * time.Second) {
// 		mem := stats.MemUsage()

// 		allocated := strconv.Itoa(int(mem.Allocated))
// 		system := strconv.Itoa(int(mem.System))

// 		if int(mem.Allocated) > 500 || int(mem.System) > 5000 {
// 			SayMention("error-tracking", "<@247905755808792580>", "> Memory usage is exceeding expectations! \n > \n > Allocated -> `"+allocated+"` \n > System -> `"+system+"`")
// 		}
// 	}
// }

func manuallyTweet(r *discordgo.MessageReactionAdd) {
	var channel string
	var message string

	// If message was sent by bot
	messageInfo, err := discord.ChannelMessage(r.ChannelID, r.MessageID)
	if err != nil {
		stats.Log(err.Error())
		return
	}
	if messageInfo.Author.ID != global.DiscordBotID {
		return
	}

	// If starts with "```Channel", get channel from message
	// Else, get channel from channel name
	if messageInfo, _ := discord.ChannelMessage(r.ChannelID, r.MessageID); strings.HasPrefix(messageInfo.Content, "```Channel") {
		s := strings.Split(strings.ReplaceAll(messageInfo.Content, "\n", " "), " ")
		channel = strings.ReplaceAll(s[1], "\n", "")
		message = strings.Join(s[3:], " ")
		message = strings.ReplaceAll(message, "```", "")
		twitter.SendTweet(channel, message)
	} else {
		c, _ := discord.Channel(r.ChannelID)
		channel = c.Name
		message = strings.ReplaceAll(messageInfo.Content, "`", "")
		twitter.SendTweet(channel, message)
	}
}

func findChannelIDs(mode string, platform string, channelName string, returnChannelID string) (platformChannelID string, discordChannelID string, success bool) {
	if mode == "add" {
		if platform == "twitch" {
			c, err := twitch.GetBroadcasterInfo(channelName)
			if err != nil {
				go SayByIDAndDelete(returnChannelID, "Is this a real twitch channel?")
				return "", "", false
			}
			platformChannelID = c.ID
		} else if platform == "youtube" {
			go SayByIDAndDelete(returnChannelID, "YouTube support not yet added.")
			return
		} else {
			go SayByIDAndDelete(returnChannelID, "Invalid platform.")
			return
		}

		c, ok := CreateDiscordChannel(channelName)
		if !ok {
			go SayByIDAndDelete(returnChannelID, "Failed to create Discord channel.")
			return "", "", false
		}
		discordChannelID = c.ID
	} else {
		for _, c := range global.Directives {
			if channelName == c.ChannelName {
				platformChannelID = c.ChannelID
				discordChannelID = c.DiscordChannelID
				break
			}
		}
	}
	return platformChannelID, discordChannelID, true
}
