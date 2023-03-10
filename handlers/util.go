package handlers

import (
	"Message-Generator/global"
	"Message-Generator/platform"
	"regexp"
	"strings"
	"time"

	"Message-Generator/markov"
)

// prepareMessageForMarkov prepares the message to be inputted into a Markov chain.
func prepareMessageForMarkov(msg platform.Message) (processed string) {
	processed = removeMentions(msg.Content)
	processed = lowercaseIfNotEmote(msg.ChannelName, processed)
	processed = removeWeirdTwitchCharactersAndTrim(processed)

	return processed
}

// removeMentions will remove usernames that are paired with @'s.
func removeMentions(message string) (processed string) {
	var s []string
	for _, word := range strings.Split(message, " ") {
		if strings.Contains(word, "@") || strings.Contains(word, global.BotName) {
			continue
		}
		s = append(s, word)
	}
	processed = strings.Join(s, " ")
	return processed
}

// lowercaseIfNotEmote takes channel and string and returns the string with everything lowercase except any emotes from that channel.
func lowercaseIfNotEmote(channel string, message string) string {
	global.EmotesMx.Lock()
	defer global.EmotesMx.Unlock()
	var new []string
	slice := strings.Split(message, " ")
	for _, word := range slice {
		match := false
		for _, emote := range global.GlobalEmotes {
			if word == emote.Name {
				match = true
				new = append(new, word)
				break
			}
		}

		if !match {
			for _, emote := range global.TwitchChannelEmotes {
				if word == emote.Name {
					match = true
					new = append(new, word)
					break
				}
			}
		}

		if !match {
			for _, c := range global.ThirdPartyChannelEmotes {
				if c.Name == channel {
					for _, emote := range c.Emotes {
						if word == emote.Name {
							match = true
							new = append(new, word)
							break
						}
					}
				}
			}
		}

		if !match {
			new = append(new, strings.ToLower(word))
		}
	}
	newMessage := strings.Join(new, " ")
	return newMessage
}

// removeWeirdTwitchCharactersAndTrim removes whitespaces that Twitch adds, such as  and 󠀀.
func removeWeirdTwitchCharactersAndTrim(message string) string {
	message = strings.ReplaceAll(message, "\x01", "")
	message = strings.ReplaceAll(message, "󠀀", "")
	slice := strings.Fields(message)
	message = strings.Join(slice, " ")
	return message
}

// checkForUrl returns if a string contains a link/url.
func checkForUrl(urlOrNot string) bool {
	r, _ := regexp.Compile(`(http|ftp|https):\/\/([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)
	return r.MatchString(urlOrNot)
}

// checkForBotUser returns if a username belongs to a bot account.
func checkForBotUser(username string) bool {
	if m, _ := regexp.MatchString(".bot", username); m {
		return true
	}
	for _, v := range global.BannedUsers {
		if v == username {
			return true
		}
	}
	return false
}

// checkForBadWording returns if a message contains a bad word or phrase.
func checkForBadWording(message string) bool {
	return global.Regex.MatchString(message)
}

// checkForCommand returns if a string is a potential command.
func checkForCommand(message string) bool {
	s := []string{"!", "%", "?", "-", ".", ",", "#", "+", "$"}
	for _, prefix := range s {
		if strings.HasPrefix(message, prefix) {
			return true
		}
	}
	return false
}

// checkForRepitition returns if a string repeats words 3 or more times.
func checkForRepitition(message string) bool {
	wordList := strings.Fields(message)
	counts := make(map[string]int)
	for _, word := range wordList {
		_, ok := counts[word]
		if ok {
			counts[word] += 1
		} else {
			counts[word] = 1
		}
	}
	for _, number := range counts {
		if number > 2 {
			return true
		}
	}
	return false
}

// passesMessageQualityCheck checks if a username or message passes the vibe check.
func passesMessageQualityCheck(username string, message string) bool {
	// Check for url
	if checkForUrl(message) {
		return false
	}

	// Check for bad wording
	if checkForBadWording(message) {
		return false
	}

	// Check usernames for bots
	if checkForBotUser(username) {
		return false
	}

	// Check for command
	if checkForCommand(message) {
		return false
	}

	// Check if message has too much repitition
	if checkForRepitition(message) {
		return false
	}

	return true
}

func mentionsBot(msg string) bool {
	return strings.Contains(strings.ToLower(msg), strings.ToLower(global.BotName))
}

// lockAPI will mark a channel as locked until the time (in seconds) has passed.
func lockAPI(timer int, channel string) bool {
	apiLocksMx.Lock()
	if apiLocks[channel] {
		apiLocksMx.Unlock()
		return false
	}
	apiLocks[channel] = true
	apiLocksMx.Unlock()
	go unlockAPI(timer, channel)
	return true
}

func unlockAPI(timer int, channel string) {
	time.Sleep(time.Duration(timer) * time.Second)
	apiLocksMx.Lock()
	apiLocks[channel] = false
	apiLocksMx.Unlock()
}

// lockDefault will mark a channel as locked until the time (in seconds) has passed.
func lockDefault(timer int, channel string) bool {
	defaultLocksMx.Lock()
	if defaultLocks[channel] {
		defaultLocksMx.Unlock()
		return false
	}
	defaultLocks[channel] = true
	defaultLocksMx.Unlock()
	go unlockDefault(timer, channel)
	return true
}

func unlockDefault(timer int, channel string) {
	time.Sleep(time.Duration(timer) * time.Second)
	defaultLocksMx.Lock()
	defaultLocks[channel] = false
	defaultLocksMx.Unlock()
}

// lockParticipation will mark a channel as locked until the time (in minutes) has passed.
func lockParticipation(time int, channel string) bool {
	participationLocksMx.Lock()
	if participationLocks[channel] {
		participationLocksMx.Unlock()
		return false
	}
	participationLocks[channel] = true
	participationLocksMx.Unlock()
	go unlockParticipation(time, channel)
	return true
}

func unlockParticipation(timer int, channel string) {
	time.Sleep(time.Duration(timer) * time.Minute)
	participationLocksMx.Lock()
	participationLocks[channel] = false
	participationLocksMx.Unlock()
}

// lockReply will mark a channel as locked until the time (in minutes) has passed.
func lockReply(time int, channel string) bool {
	replyLocksMx.Lock()
	if replyLocks[channel] {
		replyLocksMx.Unlock()
		return false
	}
	replyLocks[channel] = true
	replyLocksMx.Unlock()
	go unlockReply(time, channel)
	return true
}

func unlockReply(timer int, channel string) {
	time.Sleep(time.Duration(timer) * time.Minute)
	replyLocksMx.Lock()
	replyLocks[channel] = false
	replyLocksMx.Unlock()
}
func isSentenceTooShort(sentence string) bool {
	// Split sentence into words
	s := strings.Split(sentence, " ")

	// If there are one to two words, 5% chance to pass
	if 0 < len(s) && len(s) < 3 {
		n := global.RandomNumber(0, 100)
		if n <= 5 {
			return true
		}
	}

	return false
}

func containsOwnName(message string) bool {
	return strings.Contains(message, global.BotName)
}

func DoesSliceContainIndex(slice []string, index int) bool {
	if len(slice) > index {
		return true
	} else {
		return false
	}
}

func removeDeterminers(content string) (target string) {
	s := strings.Split(clearNonAlphanumeric(content), " ")
	ns := []string{}

	wordsToAvoid := []string{global.BotName, "me", "are", "to", "you", "i", "is", "a", "an", "the", "my", "your", "it", "its", "their", "much", "many", "of", "some", "any", "from", "such"}

wordLoop:
	for _, word := range s {
		for _, determiner := range wordsToAvoid {
			if word == determiner {
				continue wordLoop
			}
		}

		if trimmed := strings.TrimSpace(word); trimmed == "" {
			continue wordLoop
		} else {
			ns = append(ns, strings.TrimSpace(word))
		}
	}

	if len(ns) == 0 {
		return ""
	}

	return global.PickRandomFromSlice(ns)
}

func clearNonAlphanumeric(str string) string {
	nonAlphanumericRegex := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
	return nonAlphanumericRegex.ReplaceAllString(str, "")
}

func questionType(content string) (questionType string) {
	yesNoWords := []string{"will", "is", "does", "do", "are", "have"}
	explanationWords := []string{"why", "how"}

	for _, q := range yesNoWords {
		if strings.HasPrefix(content, q) {
			return "yes no question"
		}
	}

	for _, q := range explanationWords {
		if strings.HasPrefix(content, q) {
			return "explanation question"
		}
	}

	return "not a question"
}

func decideWhichChannelToUse(directive global.Directive) string {
	if directive.Settings.WhichChannelsToUse == "self" && directive.Settings.IsCollectingMessages {
		return directive.ChannelName
	}

	if directive.Settings.WhichChannelsToUse == "custom" && len(directive.Settings.CustomChannelsToUse) > 0 {
		return global.PickRandomFromSlice(directive.Settings.CustomChannelsToUse)
	}

	if directive.Settings.WhichChannelsToUse == "all" {
		return global.PickRandomFromSlice(markov.Chains())
	}

	// At this point, "directive.Settings.WhichChannelsToUse" will be except self
	var s []string
	for _, chain := range markov.Chains() {
		if chain == directive.ChannelName {
			continue
		}
		s = append(s, chain)
	}
	return global.PickRandomFromSlice(s)
}
