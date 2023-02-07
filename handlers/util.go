package handlers

import (
	"Message-Generator/discord"
	"Message-Generator/global"
	"Message-Generator/platform"
	"Message-Generator/platform/twitch"
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
	message = strings.ReplaceAll(message, "", "")
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
	if strings.Contains(username, "bot") {
		return true
	}
	for _, v := range global.BannedUsers {
		if strings.Contains(username, v) {
			return true
		}
	}
	return false
}

// checkForBadWording returns if a message contains a bad word or phrase.
func checkForBadWording(message string) bool {
	if global.Regex.MatchString(message) {
		return true
	}
	return false
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
	if strings.Contains(message, global.BotName) {
		return true
	}

	return false
}

func DoesSliceContainIndex(slice []string, index int) bool {
	if len(slice) > index {
		return true
	} else {
		return false
	}
}

// GetRandomChannel will return a random channel, unless mode == "except self", in which case channel won't be included.
func GetRandomChannel(mode string, channel string) (randomChannel string) {
	var s []string

	chains := markov.CurrentWorkers()
	for _, chain := range chains {
		if mode == "except self" && chain == channel {
			continue
		}
		s = append(s, chain)
	}

	return global.PickRandomFromSlice(s)
}

func findChannelIDs(mode string, platform string, channelName string, returnChannelID string) (platformChannelID string, discordChannelID string, success bool) {
	if mode == "add" {
		if platform == "twitch" {
			c, err := twitch.GetBroadcasterInfo(channelName)
			if err != nil {
				go discord.SayByIDAndDelete(returnChannelID, "Is this a real twitch channel?")
				return "", "", false
			}
			platformChannelID = c.ID
		} else if platform == "youtube" {
			go discord.SayByIDAndDelete(returnChannelID, "YouTube support not yet added.")
			return
		} else {
			go discord.SayByIDAndDelete(returnChannelID, "Invalid platform.")
			return
		}

		c, ok := discord.CreateDiscordChannel(channelName)
		if !ok {
			go discord.SayByIDAndDelete(returnChannelID, "Failed to create Discord channel.")
			return "", "", false
		}
		discordChannelID = c.ID
	} else {
		for _, c := range global.Directives {
			if channelName == c.ChannelName {
				platformChannelID = c.ChannelID
				discordChannelID = c.DiscordChannelID
				break
			}
		}
	}
	return platformChannelID, discordChannelID, true
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

		w := strings.ReplaceAll(word, ".", "")
		w = strings.ReplaceAll(word, ",", "")
		w = strings.ReplaceAll(word, "!", "")
		w = strings.ReplaceAll(word, "?", "")

		ns = append(ns, w)
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

	// At this point, "directive.Settings.WhichChannelsToUse" will be either all or except self
	return GetRandomChannel(directive.Settings.WhichChannelsToUse, directive.ChannelName)
}
