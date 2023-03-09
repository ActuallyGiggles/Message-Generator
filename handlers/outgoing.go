package handlers

import (
	"Message-Generator/discord"
	"Message-Generator/global"
	"Message-Generator/markov"
	"Message-Generator/platform"
	"Message-Generator/platform/twitch"
	"Message-Generator/print"
	"Message-Generator/twitter"
	"fmt"
	"strings"
	"sync"
)

var (
	currentlyMakingDefaultSentence sync.Mutex
	defaultLocks                   = make(map[string]bool)
	defaultLocksMx                 sync.Mutex
	apiLocks                       = make(map[string]bool)
	apiLocksMx                     sync.Mutex
	participationLocks             = make(map[string]bool)
	participationLocksMx           sync.Mutex
	replyLocks                     = make(map[string]bool)
	replyLocksMx                   sync.Mutex
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

// CreateDefaultSentence outputs a likely sentence to a Discord channel.
func CreateDefaultSentence(msg platform.Message) {
	if !currentlyMakingDefaultSentence.TryLock() {
		return
	}
	defer currentlyMakingDefaultSentence.Unlock()

	// Allow passage if not currently timed out.
	if !lockDefault(600, msg.ChannelName) {
		return
	}

	var recursionLimit = 5
	var timesRecursed = 0

recurse:
	target := removeDeterminers(msg.Content)
	if target == "" {
		return
	}

	oi := markov.OutputInstructions{
		Chain:  msg.ChannelName,
		Method: "TargetedMiddle",
		Target: target,
	}

	// Get output.
	output, err := markov.Out(oi)

	if err != nil {
		if timesRecursed > recursionLimit {
			// If simply not found in chain or chain is too small, ignore error.
			if strings.Contains(err.Error(), "does not exist in chain") || strings.Contains(err.Error(), "does not contain parents that match") || strings.Contains(err.Error(), "is not found in directory") {
				return
			}

			// Report if too many errors.
			print.Warning("Could not create default sentence.\nError: " + err.Error())
			return
		}

		// Recurse.
		timesRecursed++
		goto recurse
	}

	if isSentenceTooShort(output) {
		// Recurse.
		timesRecursed++
		goto recurse
	}

	if containsOwnName(output) {
		// Recurse.
		timesRecursed++
		goto recurse
	}

	OutgoingHandler("default", msg.ChannelName, "", oi, output, "")
}

// CreateAPISentence outputs a likely sentence for the API.
func CreateAPISentence(channel string) (output string, success bool) {
	// Allow passage if not currently timed out.
	if !lockAPI(1, channel) {
		return "", false
	}

	var recursionLimit = 500
	var timesRecursed = 0

recurse:
	oi := markov.OutputInstructions{
		Chain:  channel,
		Method: "RandomMiddle",
	}

	// Get output.
	output, err := markov.Out(oi)

	if err != nil {
		if timesRecursed > recursionLimit {
			// If simply not found in chain or chain is too small, ignore error.
			if strings.Contains(err.Error(), "does not exist in chain") || strings.Contains(err.Error(), "does not contain parents that match") || strings.Contains(err.Error(), "is not found in directory") {
				return
			}

			// Report if too many errors.
			print.Warning("Could not create API sentence.\nError: " + err.Error())
			return "", false
		}

		// Recurse.
		timesRecursed++
		goto recurse
	}

	if isSentenceTooShort(output) {
		timesRecursed++
		goto recurse
	}

	if containsOwnName(output) {
		timesRecursed++
		goto recurse
	}

	OutgoingHandler("api", channel, "", oi, output, "")

	return output, true
}

// CreateParticipationSentence takes in a message and outputs a targeted sentence without reply a user.
func CreateParticipationSentence(msg platform.Message, directive global.Directive) {
	// Allow passage if allowed to participate in chat.
	if !directive.Settings.Participation.IsEnabled {
		return
	}

	isLive := twitch.IsChannelLive(directive.ChannelName)

	// Allow passage if channel is online and online is enabled.
	if isLive && !directive.Settings.Participation.IsAllowedWhenOnline {
		return
	}

	// Allow passage if channel is offline and offline is enabled.
	if !isLive && !directive.Settings.Participation.IsAllowedWhenOffline {
		return
	}

	// // Allow passage if random rejection of 10% allows.
	// if randomChance := global.RandomNumber(0, 100); randomChance > 10 {
	// 	return
	// }

	// Allow passage if not currently timed out.
	if isLive {
		if !lockParticipation(directive.Settings.Participation.OnlineTimeToWait, msg.ChannelName) {
			return
		}
	} else {
		if !lockParticipation(directive.Settings.Participation.OfflineTimeToWait, msg.ChannelName) {
			return
		}
	}

	// Try each chain at least 2 times
	recursionLimit := len(markov.CurrentWorkers())
	timesRecursed := 0

recurse:
	target := removeDeterminers(msg.Content)
	if target == "" {
		return
	}

	oi := markov.OutputInstructions{
		Chain:  decideWhichChannelToUse(directive),
		Method: "TargetedMiddle",
		Target: target,
	}

	// Get output.
	output, err := markov.Out(oi)

	// Handle error.
	if err != nil {
		if strings.Contains(err.Error(), "Target is empty") {
			return
		}

		// Stop if too much recursing.
		if timesRecursed > recursionLimit {
			// If simply not found in chain or chain is too small, ignore error.
			if strings.Contains(err.Error(), "does not exist in chain") || strings.Contains(err.Error(), "does not contain parents that match") || strings.Contains(err.Error(), "is not found in directory") {
				return
			}

			// Report if too many errors.
			print.Warning("Could not create participation sentence.\nTrigger Message: " + msg.Content + "\n" + "Error: " + err.Error())
			return
		}

		// Recurse.
		timesRecursed++
		goto recurse
	}

	if isSentenceTooShort(output) {
		timesRecursed++
		goto recurse
	}

	if containsOwnName(output) {
		timesRecursed++
		goto recurse
	}

	// Handle output.
	OutgoingHandler("participation", msg.ChannelName, msg.Content, oi, output, "")
}

// CreateReplySentence takes in a message and outputs a targeted sentence that directly mentions a user.
func CreateReplySentence(msg platform.Message, directive global.Directive) {
	// If not allowed to respond to mentions, return.
	if !directive.Settings.Reply.IsEnabled {
		return
	}

	isLive := twitch.IsChannelLive(directive.ChannelName)

	fmt.Println(directive.ChannelName, "is live:", isLive)

	// Allow passage if channel is online and online is enabled.
	if isLive && !directive.Settings.Reply.IsAllowedWhenOnline {
		return
	}

	// Allow passage if channel is offline and offline is enabled.
	if !isLive && !directive.Settings.Reply.IsAllowedWhenOffline {
		return
	}

	// Allow passage if not currently timed out. Also, based on if offline or online.
	if isLive {
		if !lockReply(directive.Settings.Reply.OnlineTimeToWait, msg.ChannelName) {
			return
		}
	} else {
		if !lockReply(directive.Settings.Reply.OfflineTimeToWait, msg.ChannelName) {
			return
		}
	}

	// Try each chain at least 2 times
	recursionLimit := len(markov.CurrentWorkers())
	timesRecursed := 0

recurse:
	var oi markov.OutputInstructions

	questionType := questionType(msg.Content)
	if questionType == "yes no question" {
		oi = markov.OutputInstructions{
			Method: "TargetedBeginning",
			Chain:  decideWhichChannelToUse(directive),
			Target: global.PickRandomFromSlice([]string{"yes", "no", "maybe", "absolutely", "absolutely", "never", "always"}),
		}
	} else {
		target := removeDeterminers(msg.Content)
		if target == "" {
			return
		}

		oi = markov.OutputInstructions{
			Method: "TargetedMiddle",
			Chain:  decideWhichChannelToUse(directive),
			Target: target,
		}
	}

	output, err := markov.Out(oi)

	// Handle error.
	if err != nil {
		if strings.Contains(err.Error(), "Target is empty") {
			return
		}

		if timesRecursed > recursionLimit {
			// If simply not found in chain or chain is too small, ignore error.
			if strings.Contains(err.Error(), "does not exist in chain") || strings.Contains(err.Error(), "does not contain parents that match") || strings.Contains(err.Error(), "is not found in directory") {
				return
			}

			// Report if too many errors.
			print.Warning("Could not create reply sentence.\nTrigger Message: " + msg.Content + "\n" + "Error: " + err.Error())
			return
		}

		// Recurse.
		timesRecursed++
		goto recurse
	}

	if isSentenceTooShort(output) {
		timesRecursed++
		goto recurse
	}

	if containsOwnName(output) {
		timesRecursed++
		goto recurse
	}

	// Handle output.
	OutgoingHandler("reply", msg.ChannelName, msg.Content, oi, output, msg.AuthorName)
}
