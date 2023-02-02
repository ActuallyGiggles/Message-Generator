package handlers

import (
	"Message-Generator/discord"
	"Message-Generator/markov"
	"Message-Generator/platform/twitch"
	"Message-Generator/twitter"
	"strings"
	"time"
)

func OutgoingHandler(origin string, channelUsed string, sendBackToChannel string, method string, message string, mention string) {
	// Say message into discord all channel and respective discord channel.
	discord.Say("all", "Channel: "+channelUsed+"\nMessage: "+message)
	discord.Say(channelUsed, message)

	// If message is three words or longer, add to potential tweets.
	// continue
	if len(strings.Split(message, " ")) >= 3 {
		twitter.AddMessageToPotentialTweets(channelUsed, message)
	}

	// If message is from api, send to website results.
	// stop
	if origin == "api" {
		discord.Say("website-results", "Channel: "+channelUsed+"\nMethod: "+method+"\nMessage: "+message)
		return
	}

	// If message is prompted by participation sentence, say to respective twitch channel.
	// stop
	if origin == "participation" {
		twitch.Say(sendBackToChannel, message)
		discord.Say("participation", "Channel Used: "+channelUsed+"\nMethod: "+method+"\nChannel Sent To: "+sendBackToChannel+"\nMessage: "+message)
		return
	}

	// If message is prompted by reply sentence, say to respective channel and say in discord reply channel.
	// stop
	if origin == "reply" {
		twitch.Say(sendBackToChannel, "@"+mention+" "+message)
		discord.Say("reply", "Channel Used: "+channelUsed+"\nMethod: "+method+"\nChannel Sent To: "+sendBackToChannel+"\nMessage: @"+mention+" "+message)
		return
	}

	return
}

func outputTicker() {
	for range time.Tick(5 * time.Minute) {
		chains := markov.CurrentWorkers()
		for _, chain := range chains {
			CreateDefaultSentence(chain)
		}
	}
}
