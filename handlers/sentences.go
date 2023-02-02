package handlers

import (
	"Message-Generator/discord"
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
	participationLocks   = make(map[string]bool)
	participationLocksMx sync.Mutex
	apiLocks             = make(map[string]bool)
	apiLocksMx           sync.Mutex
)

// CreateDefaultSentence outputs a likely sentence to a Discord channel.
func CreateDefaultSentence(channel string) {
	// Allow passage if not currently timed out.
	if !lockDefault(300, channel) {
		return
	}

	var recursionLimit = 5
	var timesRecursed = 0

recurse:
	// Get output.
	output, err := markov.Out(markov.OutputInstructions{
		Chain:  channel,
		Method: "LikelyBeginning",
	})

	if err != nil {
		if timesRecursed > recursionLimit {
			// Report if too many errors.
			print.Warning("Could not create default sentence.\nError: " + err.Error())
			discord.Say("error-tracking", err.Error())

			return
		}

		// Recurse.
		timesRecursed++
		goto recurse
	}

	if isSentenceTooShort(output) {
		return
	}

	if containsOwnName(output) {
		return
	}

	OutgoingHandler("default", channel, channel, "LikelyBeginning", output, "")
}

// CreateAPISentence outputs a likely sentence for the API.
func CreateAPISentence(channel string) (output string, success bool) {
	// Allow passage if not currently timed out.
	if !lockAPI(1, channel) {
		return "", false
	}

	var recursionLimit = 100
	var timesRecursed = 0

recurse:
	method := global.PickRandomFromSlice([]string{"LikelyBeginning", "LikelyEnding"})
	// Get output.
	output, err := markov.Out(markov.OutputInstructions{
		Chain:  channel,
		Method: method,
	})

	if err != nil {
		if timesRecursed > recursionLimit {
			// Report if too many errors.
			print.Warning("Could not create API sentence.\nError: " + err.Error())
			discord.Say("error-tracking", err.Error())

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

	OutgoingHandler("api", channel, channel, method, output, "")

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

	// Allow passage if random rejection of 50% allows.
	if randomChance := global.RandomNumber(0, 100); randomChance > 50 {
		return
	}

	// Allow passage if not currently timed out.
	if !lockParticipation(global.RandomNumber(1, 10), msg.ChannelName) {
		return
	}

	var recursionLimit = 50
	var timesRecursed = 0

recurse:
	chainToUse := decideWhichChannelToUse(directive)

	// Pick a method and target word.
	method := global.PickRandomFromSlice([]string{"TargetedBeginning", "TargetedMiddle", "TargetedEnding"})
	target := removeDeterminers(strings.ReplaceAll(msg.Content, ".", ""))

	// Get output.
	output, err := markov.Out(markov.OutputInstructions{
		Chain:  chainToUse,
		Method: method,
		Target: target,
	})

	// Handle error.
	if err != nil {
		// Stop if too much recursing.
		if timesRecursed > recursionLimit {
			// Report if too many errors.
			print.Warning("Could not create participation sentence.\nTrigger Message:\n\t" + msg.Content + "\n" + "Error: " + err.Error())
			discord.Say("error-tracking", err.Error())
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
	OutgoingHandler("participation", chainToUse, msg.ChannelName, method, output, "")

	return
}

// CreateReplySentence takes in a message and outputs a targeted sentence that directly mentions a user.
func CreateReplySentence(msg platform.Message, directive global.Directive) {
	// If does not mention, return.
	if !strings.Contains(strings.ToLower(msg.Content), strings.ToLower(global.BotName)) {
		return
	}

	// If not allowed to respond to mentions, return.
	if !directive.Settings.Reply.IsEnabled {
		return
	}

	// Allow passage if channel is online and online is enabled or if channel is offline and offline is enabled.
	if (twitch.IsChannelLive(directive.ChannelName) && !directive.Settings.Reply.IsAllowedWhenOnline) || (!twitch.IsChannelLive(directive.ChannelName) && !directive.Settings.Reply.IsAllowedWhenOffline) {
		return
	}

	recursionLimit := len(markov.CurrentWorkers())
	timesRecursed := 0

recurse:
	chainToUse := decideWhichChannelToUse(directive)

	var method string
	var target string

	questionType := questionType(msg.Content)
	if questionType == "yes no question" {
		method = "TargetedBeginning"
		target = global.PickRandomFromSlice([]string{"yes", "no", "maybe", "absolutely", "absolutely", "who knows"})
	} else if questionType == "explanation question" {
		method = "TargetedBeginning"
		target = global.PickRandomFromSlice([]string{"because", "idk", "idc", "well", "you see"})
	} else {
		method = global.PickRandomFromSlice([]string{"TargetedBeginning", "TargetedMiddle", "TargetedEnding"})
		target = removeDeterminers(strings.ReplaceAll(msg.Content, ".", ""))
		if target == "" {
			return
		}
	}

	output, err := markov.Out(markov.OutputInstructions{
		Method: method,
		Chain:  chainToUse,
		Target: target,
	})

	// Handle error.
	if err != nil {
		if timesRecursed > recursionLimit {
			// Report if too many errors.
			print.Warning("Could not create reply sentence.\nTrigger Message:\n\t" + msg.Content + "\n" + "Error: " + err.Error())
			discord.Say("error-tracking", err.Error())
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
	OutgoingHandler("reply", chainToUse, msg.ChannelName, method, output, msg.AuthorName)

	return
}
