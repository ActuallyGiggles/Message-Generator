package handlers

import (
	"Message-Generator/discord"
	"Message-Generator/global"
	"Message-Generator/markov"
	"Message-Generator/platform"
	"fmt"
)

func Incoming(c chan platform.Message) {
	for msg := range c {
		if !passesMessageQualityCheck(msg.AuthorName, msg.Content) {
			continue
		}

		preparedMessage := prepareMessageForMarkov(msg)

		var exists bool

		for _, directive := range global.Directives {
			if directive.ChannelName == msg.ChannelName {
				exists = true

				if directive.Settings.IsCollectingMessages {
					go markov.In(msg.ChannelName, preparedMessage)
					go CreateDefaultSentence(msg.ChannelName)
				}

				go CreateReplySentence(msg, directive)

				msg.Content = preparedMessage
				go CreateParticipationSentence(msg, directive)
			}
		}

		if !exists {
			discord.Say("error-tracking", fmt.Sprintf("%s does not exist as a directive", msg.ChannelName))
		}

		continue
	}
}
