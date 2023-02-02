package discord

import (
	"Message-Generator/global"
	"Message-Generator/print"
	"strings"

	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	discord *discordgo.Session
)

func Start(errorChannel chan string) {
	bot, err := discordgo.New("Bot " + global.DiscordToken)
	discord = bot
	if err != nil {
		panic(err)
	}

	err = discord.Open()
	if err != nil {
		panic(err)
	}

	CollectSupportDiscordChannelIDs(discord)
	go reportErrors(errorChannel)

	discord.AddHandler(messageHandler)
	discord.AddHandler(reactionHandler)
}

// messageHandler receives messages and sends them into the in channel.
func messageHandler(session *discordgo.Session, message *discordgo.MessageCreate) {
	messageSlice := strings.Split(strings.TrimPrefix(message.Content, global.Prefix), " ")

	// If not by admin, return
	if message.Author.ID != global.DiscordOwnerID {
		return
	}

	// If dialogue is ongoing, send to dialogue
	if dialogueChannel != nil {
		dialogueChannel <- Dialogue{Arguments: messageSlice, MessageID: message.ID}
		return
	}

	// If does not have prefix, return
	if !strings.HasPrefix(message.Content, global.Prefix) {
		return
	}

	commandsHandler(IncomingMessage{
		ChannelID:  message.ChannelID,
		AuthorName: message.Author.Username,
		MessageID:  message.ID,
		Command:    messageSlice[0],
		Args:       messageSlice[1:],
	})

}

func reactionHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	// If correct emoji and correct user
	if r.UserID == global.DiscordOwnerID && r.Emoji.Name == global.DiscordTweetEmote {
		manuallyTweet(r)
	}
}

// SayByID sends a message to a Discord channel via the channel ID. Returns the message ID.
func SayByID(channelId string, message string) (id *discordgo.Message) {
	m, err := discord.ChannelMessageSend(channelId, WrapInCodeBlock(message))
	if err != nil {
		print.Error("SayById failed \n" + err.Error())
		return
	}
	return m
}

// Say sends a message to a Discord channel if the channel ID is already recorded for that channel name.
func Say(channel string, message string) {
	var sendToChannel string

	switch channel {
	case "all":
		sendToChannel = global.DiscordAllChannelID
	case "reply":
		sendToChannel = global.DiscordReplyChannelID
	case "participation":
		sendToChannel = global.DiscordParticipationChannelID
	case "error-tracking":
		sendToChannel = global.DiscordErrorTrackingChannelID
	case "website-results":
		sendToChannel = global.DiscordWebsiteResultsChannelID
	default:
		for _, directive := range global.Directives {
			if directive.ChannelName == channel {
				sendToChannel = directive.DiscordChannelID
			}
		}
	}

	SayByID(sendToChannel, message)
	return
}

// SayMention sends a reply message to a specific user via Say.
func SayMention(channel string, mention string, message string) {
	content := mention + "\n" + message
	Say(channel, content)
}

// SayByIDAndDelete uses SayById, but then deletes the message after 15 seconds.
func SayByIDAndDelete(channelID string, message string) {
	m := SayByID(channelID, message)
	time.Sleep(time.Duration(15) * time.Second)
	DeleteDiscordMessage(channelID, m.ID)
}

// GetChannels returns a list of Discord channels connected to.
func GetChannels(session *discordgo.Session) (channels []*discordgo.Channel, err error) {
	s, err := session.GuildChannels(global.DiscordGuildID)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateDiscordChannel creates a text channel in Discord by the passed name. Returns the channel, if the function was successful.
func CreateDiscordChannel(name string) (channel *discordgo.Channel, ok bool) {
	c, err := discord.GuildChannelCreate(global.DiscordGuildID, name, discordgo.ChannelTypeGuildText)
	if err != nil {
		print.Error("CreateDiscordChannel failed\n" + err.Error())
		return nil, false
	}
	return c, true
}

// DeleteDiscordChannel deletes any text channel in Discord by the passed name. Returns if the function was successful.
func DeleteDiscordChannel(name string) (ok bool) {
	for _, c := range global.Directives {
		if c.ChannelName == name {
			_, err := discord.ChannelDelete(c.ChannelID)
			if err != nil {
				print.Error("DeleteDiscordChannel failed\n" + err.Error())
			}
		}
	}
	return true
}

func DeleteDiscordMessage(channelID string, messageID string) {
	err := discord.ChannelMessageDelete(channelID, messageID)
	if err != nil {
		print.Error("DeleteDiscordMessage failed\n" + err.Error())
	}
}

// CollectSupportDiscordChannelIDs will collect channel IDs for channels that are necessary to send to but do not have platform channel counterpart.
func CollectSupportDiscordChannelIDs(session *discordgo.Session) (ok bool) {
	channels, err := GetChannels(session)
	if err != nil {
		panic(err)
	}

	for _, channel := range channels {
		channel = *&channel
		if channel.Name == "all" {
			global.DiscordAllChannelID = channel.ID
		}

		if channel.Name == "participation" {
			global.DiscordParticipationChannelID = channel.ID
		}

		if channel.Name == "reply" {
			global.DiscordReplyChannelID = channel.ID
		}

		if channel.Name == "error-tracking" {
			global.DiscordErrorTrackingChannelID = channel.ID
		}

		if channel.Name == "website-results" {
			global.DiscordWebsiteResultsChannelID = channel.ID
		}
	}

	return true
}

func reportErrors(errorChannel chan string) {
	for err := range errorChannel {
		Say("error-tracking", err)
	}
}
