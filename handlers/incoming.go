package handlers

import (
	"Message-Generator/discord"
	"Message-Generator/global"
	"Message-Generator/markov"
	"Message-Generator/platform"
	"fmt"
	"strings"
)

func Incoming(c chan platform.Message) {
	for msg := range c {
		if !passesMessageQualityCheck(msg.AuthorName, msg.Content) {
			continue
		}

		preparedContent := prepareMessageForMarkov(msg)

		var exists bool

		for _, directive := range global.Directives {
			if directive.ChannelName == msg.ChannelName {
				exists = true

				if directive.Settings.IsCollectingMessages {
					go markov.In(msg.ChannelName, preparedContent)
					go CreateDefaultSentence(msg.ChannelName)
				}

				preparedMsg := msg
				preparedMsg.Content = preparedContent

				// If message contains a ping for the bot, run a reply
				if strings.Contains(strings.ToLower(msg.Content), strings.ToLower(global.BotName)) {
					go CreateReplySentence(preparedMsg, directive)
				}

				go CreateParticipationSentence(preparedMsg, directive)
			}
		}

		if !exists {
			discord.Say("error-tracking", fmt.Sprintf("%s does not exist as a directive", msg.ChannelName))
		}

		continue
	}
}
