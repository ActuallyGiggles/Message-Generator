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
		go func(msg platform.Message) {
			if !passesMessageQualityCheck(msg.AuthorName, msg.Content) {
				return
			}

			for _, directive := range global.Directives {
				if directive.ChannelName == msg.ChannelName {
					if mentionsBot(msg.Content) {
						go CreateReplySentence(msg, directive)
					} else {
						go CreateParticipationSentence(msg, directive)
					}

					msg.Content = prepareMessageForMarkov(msg)

					if directive.Settings.IsCollectingMessages {
						go markov.In(msg.ChannelName, msg.Content)
						go CreateDefaultSentence(msg)
					}

					return
				}
			}

			discord.Say("error-tracking", fmt.Sprintf("%s does not exist as a directive", msg.ChannelName))
		}(msg)
	}
}
