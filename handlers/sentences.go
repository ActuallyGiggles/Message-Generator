package handlers

import (
	"Message-Generator/global"
	"Message-Generator/platform"
	"Message-Generator/platform/twitch"
	"Message-Generator/print"
	"strings"
	"sync"

	"Message-Generator/markov"
)

var (
	defaultLocks         = make(map[string]bool)
	defaultLocksMx       sync.Mutex
	apiLocks             = make(map[string]bool)
	apiLocksMx           sync.Mutex
	participationLocks   = make(map[string]bool)
	participationLocksMx sync.Mutex
	replyLocks           = make(map[string]bool)
	replyLocksMx         sync.Mutex
)

// CreateDefaultSentence outputs a likely sentence to a Discord channel.
func CreateDefaultSentence(msg platform.Message) {
	// Allow passage if not currently timed out.
	if !lockDefault(300, msg.ChannelName) {
		return
	}

	var recursionLimit = 50
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
			// If simply not found in chain, ignore error.
			if strings.Contains(err.Error(), "does not exist in chain") || strings.Contains(err.Error(), "does not contain parents that match") {
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
			// If simply not found in chain, ignore error.
			if strings.Contains(err.Error(), "does not exist in chain") || strings.Contains(err.Error(), "does not contain parents that match") {
				return "", false
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

	// Allow passage if channel is online and online is enabled or if channel is offline and offline is enabled.
	if (twitch.IsChannelLive(directive.ChannelName) && !directive.Settings.Participation.IsAllowedWhenOnline) || (!twitch.IsChannelLive(directive.ChannelName) && !directive.Settings.Participation.IsAllowedWhenOffline) {
		return
	}

	// Allow passage if random rejection of 10% allows.
	if randomChance := global.RandomNumber(0, 100); randomChance > 10 {
		return
	}

	// Allow passage if not currently timed out.
	if !lockParticipation(global.RandomNumber(5, 30), msg.ChannelName) {
		return
	}

	// Try each chain at least 2 times
	recursionLimit := len(markov.CurrentWorkers()) * 2
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

	// Allow passage if channel is online and online is enabled or if channel is offline and offline is enabled.
	if (twitch.IsChannelLive(directive.ChannelName) && !directive.Settings.Reply.IsAllowedWhenOnline) || (!twitch.IsChannelLive(directive.ChannelName) && !directive.Settings.Reply.IsAllowedWhenOffline) {
		return
	}

	// Allow passage if not currently timed out.
	if !lockReply(global.RandomNumber(0, 1), msg.ChannelName) {
		return
	}

	// Try each chain at least 2 times
	recursionLimit := len(markov.CurrentWorkers()) * 2
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
