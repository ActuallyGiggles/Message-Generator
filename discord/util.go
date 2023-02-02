package discord

import (
	"Message-Generator/global"
	"Message-Generator/platform/twitch"
	"Message-Generator/print"
	"Message-Generator/twitter"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// wrapInCodeBlock returns a message wrapped in three back ticks for Discord formatting.
func WrapInCodeBlock(message string) string {
	return "```" + message + "```"
}

func manuallyTweet(r *discordgo.MessageReactionAdd) {
	// If message was sent by bot
	messageInfo, err := discord.ChannelMessage(r.ChannelID, r.MessageID)
	if err != nil {
		print.Error(err.Error())
		return
	}

	if messageInfo.Author.ID != global.DiscordBotID {
		return
	}

	var channel string
	var message string

	// If starts with "```Channel:", get channel and message
	// If starts with "```Channel Used:", get channel and message differently
	// Else, get channel from channel name and message from content
	if strings.HasPrefix(messageInfo.Content, "```Channel:") {
		s := strings.Split(strings.ReplaceAll(strings.ReplaceAll(messageInfo.Content, "`", ""), "\n", " "), " ")
		channel = s[1]
		message = strings.Join(s[3:], " ")
	} else if strings.HasPrefix(messageInfo.Content, "```Channel Used:") {
		s := strings.Split(strings.ReplaceAll(strings.ReplaceAll(messageInfo.Content, "`", ""), "\n", " "), " ")
		channel = s[2]
		message = strings.Join(s[8:], " ")
	} else {
		c, _ := discord.Channel(r.ChannelID)
		channel = c.Name
		message = strings.ReplaceAll(messageInfo.Content, "`", "")
	}

	twitter.SendTweet(channel, message)
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

func (IDs *MessageIDs) add(ID string) {
	IDs.IDs = append(IDs.IDs, ID)
}

func (IDs *MessageIDs) delete(channelID string) {
	for _, mID := range IDs.IDs {
		DeleteDiscordMessage(channelID, mID)
	}
}
