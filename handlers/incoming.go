package handlers

import (
	"Message-Generator/discord"
	"Message-Generator/global"
	"Message-Generator/markov"
	"Message-Generator/platform"
	"fmt"
)

func Incoming(c chan platform.Message) {
messageLoop:
	for msg := range c {
		if !passesMessageQualityCheck(msg.AuthorName, msg.Content) {
			continue
		}

		for _, directive := range global.Directives {
			if directive.ChannelName == msg.ChannelName {
				mentionsBot := mentionsBot(msg.Content)
				msg.Content = prepareMessageForMarkov(msg)

				if directive.Settings.IsCollectingMessages {
					go markov.In(msg.ChannelName, msg.Content)
					go CreateDefaultSentence(msg)
				}

				if mentionsBot {
					go CreateReplySentence(msg, directive)
				} else {
					go CreateParticipationSentence(msg, directive)
				}

				continue messageLoop
			}
		}

		discord.Say("error-tracking", fmt.Sprintf("%s does not exist as a directive", msg.ChannelName))
	}
}
