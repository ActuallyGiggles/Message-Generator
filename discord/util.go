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

	messageInfo.Content = strings.ReplaceAll(strings.ReplaceAll(messageInfo.Content, "`", ""), "\n", " ")

	switch r.ChannelID {
	case global.DiscordReplyChannelID, global.DiscordParticipationChannelID:
		channel = getStringInBetween(messageInfo.Content, "Channel Used:", "Method:")
		message = getStringToEnd(messageInfo.Content, "Message:")
	case global.DiscordAllChannelID, global.DiscordWebsiteResultsChannelID:
		channel = getStringInBetween(messageInfo.Content, "Channel:", "Method:")
		message = getStringToEnd(messageInfo.Content, "Message:")
	default:
		c, _ := discord.Channel(r.ChannelID)
		channel = c.Name
		message = messageInfo.Content
	}

	twitter.SendTweet(channel, message)
}

// getStringInBetween Returns empty string if no start string found
func getStringInBetween(str, start, end string) (result string) {
	s := strings.Index(str, start)
	if s == -1 {
		return result
	}
	new := str[s+len(start):]
	e := strings.Index(new, end)
	if e == -1 {
		return result
	}
	return strings.TrimSpace(new[:e])
}

// getStringToEnd Returns empty string if no start string found
func getStringToEnd(str, start string) (result string) {
	s := strings.Index(str, start)
	if s == -1 {
		return
	}
	return strings.TrimSpace(str[s+len(start):])
}

func findChannelIDs(mode, platform, channelName, returnChannelID string) (platformChannelID, discordChannelID string, success bool) {
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
