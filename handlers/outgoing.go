package handlers

import (
	"Message-Generator/discord"
	"Message-Generator/markov"
	"Message-Generator/platform/twitch"
	"Message-Generator/twitter"
	"strings"
)

func OutgoingHandler(origin string, sendBackToChannel string, triggerSentence string, oi markov.OutputInstructions, message string, mention string) {
	// Say message into discord all channel and respective discord channel.
	discord.Say("all", "Channel: "+oi.Chain+"\nMessage: "+message)
	discord.Say(oi.Chain, message)

	// If message is three words or longer, add to potential tweets.
	// continue
	if len(strings.Split(message, " ")) >= 3 {
		twitter.AddMessageToPotentialTweets(oi.Chain, message)
	}

	// If message is from api, send to website results.
	// stop
	if origin == "api" {
		discord.Say("website-results", "Channel: "+oi.Chain+"\nMessage: "+message)
		return
	}

	// If message is prompted by participation sentence, say to respective twitch channel.
	// stop
	if origin == "participation" {
		twitch.Say(sendBackToChannel, message)
		discord.Say("participation", "Channel Sent To: "+sendBackToChannel+"\nChannel Used: "+oi.Chain+"\nMethod: "+oi.Method+"\nTarget: "+oi.Target+"\nTrigger Sentence: "+triggerSentence+"\nMessage: "+message)
		return
	}

	// If message is prompted by reply sentence, say to respective channel and say in discord reply channel.
	// stop
	if origin == "reply" {
		twitch.Say(sendBackToChannel, "@"+mention+" "+message)
		discord.Say("reply", "Channel Sent To: "+sendBackToChannel+"\nChannel Used: "+oi.Chain+"\nMethod: "+oi.Method+"\nTarget: "+oi.Target+"\nTrigger Sentence: "+triggerSentence+"\nMessage: @"+mention+" "+message)
		return
	}
}
