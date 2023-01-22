package handlers

import (
	"Twitch-Message-Generator/discord"
	"Twitch-Message-Generator/markov"
	"Twitch-Message-Generator/platform/twitch"
	"Twitch-Message-Generator/twitter"
	"strings"
	"time"
)

func OutgoingHandler(origin string, channelUsed string, sendBackToChannel string, message string, mention string) {
	str := "Channel: " + channelUsed + "\nMessage: " + message

	// Say message into discord all channel and respective discord channel.
	discord.Say("all", str)
	discord.Say(channelUsed, message)

	// If message is three words or longer, add to potential tweets.
	// continue
	if len(strings.Split(message, " ")) >= 3 {
		twitter.AddMessageToPotentialTweets(channelUsed, message)
	}

	// If message is from api, send to website results.
	// stop
	if origin == "api" {
		discord.Say("website-results", str)
		return
	}

	// If message is prompted by participation sentence, say to respective twitch channel.
	// stop
	if origin == "participation" {
		twitch.Say(sendBackToChannel, message)
		discord.Say("participation", "Channel Used: "+channelUsed+"\nChannel Sent To: "+sendBackToChannel+"\nMessage: "+message)
		return
	}

	// If message is prompted by reply sentence, say to respective channel and say in discord reply channel.
	// stop
	if origin == "reply" {
		twitch.Say(sendBackToChannel, "@"+mention+" "+message)
		discord.Say("reply", "Channel Used: "+channelUsed+"\nChannel Sent To: "+sendBackToChannel+"\nMessage: @"+mention+" "+message)
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
