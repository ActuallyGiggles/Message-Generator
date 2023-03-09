package discord

import (
	"Message-Generator/global"
	"Message-Generator/markov"
	"Message-Generator/platform/twitch"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

var (
	dialogueChannel chan Dialogue
)

// commandsHandler receives commands from an admin and returns a response.
func commandsHandler(message IncomingMessage) {
	switch message.Command {
	// Directives settings
	case "showchannels":
		showChannels(message.ChannelID, message.MessageID)
	case "showchanneldetailed":
		showDirectiveDetailed(message.ChannelID, message.MessageID, message.Args)
	case "addchannel":
		addDirective(message.ChannelID, message.MessageID)
	case "updatechannel":
		updateDirective(message.ChannelID, message.MessageID)
	case "removechannel":
		removeDirective(message.ChannelID, message.MessageID, message.Args)

	// Resources settings
	case "showregex":
		showRegex(message.ChannelID, message.MessageID)
	case "addregex":
		addRegex(message.ChannelID, message.MessageID, message.Args)
	case "removeregex":
		removeRegex(message.ChannelID, message.MessageID, message.Args)

	case "showbannedusers":
		showBannedUsers(message.ChannelID, message.MessageID)
	case "addbanneduser":
		addBannedUser(message.ChannelID, message.MessageID, message.Args)
	case "removebanneduser":
		removeBannedUser(message.ChannelID, message.MessageID, message.Args)

		// Misc
	case "cleanse":
		cleanse(message.ChannelID, message.MessageID, message.Args)
	case "help":
		help(message.ChannelID, message.MessageID)
	}
}

func showChannels(channelID string, messageID string) {
	defer DeleteDiscordMessage(channelID, messageID)
	var channels []string
	for _, directive := range global.Directives {
		channels = append(channels, directive.ChannelName)
	}
	SayByIDAndDelete(channelID, strings.Join(channels, ",\n"))
}

func addDirective(channelID string, messageID string) {
	var conversationIDs MessageIDs

	defer func() {
		conversationIDs.delete(channelID)
		dialogueChannel = nil
	}()

	dialogueChannel = make(chan Dialogue)
	conversationIDs.add(messageID)
	channel := global.Directive{}

getThePlatform:
	// Get platform
	conversationIDs.add(SayByID(channelID, "What is the platform?\n(1) Twitch\n(2) Youtube").ID)
	platform := <-dialogueChannel
	conversationIDs.add(platform.MessageID)
	switch platform.Arguments[0] {
	default:
		conversationIDs.add(SayByID(channelID, "Not a proper platform").ID)
		goto getThePlatform
	case "cancel":
		return
	case "1":
		channel.Platform = "twitch"
	case "2":
		channel.Platform = "youtube"
	}

	// Get channel name
	conversationIDs.add(SayByID(channelID, "What is the channel name?").ID)
	channelName := <-dialogueChannel
	conversationIDs.add(channelName.MessageID)
	switch channelName.Arguments[0] {
	case "cancel":
		return
	default:
		channel.ChannelName = strings.ToLower(channelName.Arguments[0])
	}

	// Return if channel is already added
	for _, c := range global.Directives {
		if c.ChannelName == channelName.Arguments[0] {
			go SayByIDAndDelete(channelID, "Channel is already added.")
			return
		}
	}

	// Get platform channel ID and channel ID
	conversationIDs.add(SayByID(channelID, "Gathering IDs...").ID)
	platformChannelID, discordChannelID, success := findChannelIDs("add", channel.Platform, channelName.Arguments[0], channelID)
	if !success {
		return
	}
	channel.ChannelID = platformChannelID
	channel.DiscordChannelID = discordChannelID

getTheSettingsToSet:
	conversationIDs.add(SayByID(channelID, "For the following options, type 0 if false and 1 if true:\n\n1. Will be collecting messages into Markov chain?\n2. Will be allowed to reply?\n3. Will be allowed to reply in online chat?\n4. Will be allowed to reply in offline chat?\n5. Will be allowed to participate with chat?\n6. Will be allowed to participate online?\n7. Will be allowed to participate offline?").ID)
	boolSettings := <-dialogueChannel
	conversationIDs.add(boolSettings.MessageID)
	if boolSettings.Arguments[0] == "cancel" {
		return
	}
	for i, setting := range boolSettings.Arguments {
		if result, err := strconv.ParseBool(setting); err != nil {
			SayByIDAndDelete(channelID, "Unable to parse settings. Cancelled.")
			DeleteDiscordChannel(channelName.Arguments[0])
			return
		} else {
			switch i {
			default:
				conversationIDs.add(SayByID(channelID, "Not a proper answer.").ID)
				goto getTheSettingsToSet
			case 0:
				channel.Settings.IsCollectingMessages = result
			case 1:
				channel.Settings.Reply.IsEnabled = result
			case 2:
				channel.Settings.Reply.IsAllowedWhenOnline = result
				if result {
					conversationIDs.add(SayByID(channelID, "Minutes to wait before replying online again?").ID)
					time := <-dialogueChannel
					conversationIDs.add(time.MessageID)
					timeParsed, success := parseTimeToWait(channelID, time.Arguments)
					if !success {
						DeleteDiscordChannel(channelName.Arguments[0])
						return
					}
					channel.Settings.Reply.OnlineTimeToWait = timeParsed
				}
			case 3:
				channel.Settings.Reply.IsAllowedWhenOffline = result
				if result {
					conversationIDs.add(SayByID(channelID, "Minutes to wait before replying offline again?").ID)
					time := <-dialogueChannel
					conversationIDs.add(time.MessageID)
					timeParsed, success := parseTimeToWait(channelID, time.Arguments)
					if !success {
						DeleteDiscordChannel(channelName.Arguments[0])
						return
					}
					channel.Settings.Reply.OfflineTimeToWait = timeParsed
				}
			case 4:
				channel.Settings.Participation.IsEnabled = result
			case 5:
				channel.Settings.Participation.IsAllowedWhenOnline = result
				if result {
					conversationIDs.add(SayByID(channelID, "Minutes to wait before participating online again?").ID)
					time := <-dialogueChannel
					conversationIDs.add(time.MessageID)
					timeParsed, success := parseTimeToWait(channelID, time.Arguments)
					if !success {
						DeleteDiscordChannel(channelName.Arguments[0])
						return
					}
					channel.Settings.Participation.OnlineTimeToWait = timeParsed
				}
			case 6:
				channel.Settings.Participation.IsAllowedWhenOffline = result
				if result {
					conversationIDs.add(SayByID(channelID, "Minutes to wait before participating offline again?").ID)
					time := <-dialogueChannel
					conversationIDs.add(time.MessageID)
					timeParsed, success := parseTimeToWait(channelID, time.Arguments)
					if !success {
						DeleteDiscordChannel(channelName.Arguments[0])
						return
					}
					channel.Settings.Participation.OfflineTimeToWait = timeParsed
				}
			}
		}
	}

getWhatChannelToUse:
	conversationIDs.add(SayByID(channelID, "What chains will this channel use to post with?\n\nAll (1)     All except self (2)     Self (3)     Custom (4)\n\nIf custom, what are the custom channels to use?").ID)
	responseSettings := <-dialogueChannel
	conversationIDs.add(responseSettings.MessageID)
	if responseSettings.Arguments[0] == "cancel" {
		return
	}
	mode := responseSettings.Arguments[0]
	customChannels := responseSettings.Arguments[1:]
	switch mode {
	default:
		conversationIDs.add(SayByID(channelID, "Not a proper answer.").ID)
		goto getWhatChannelToUse
	case "1", "all", "All":
		channel.Settings.WhichChannelsToUse = "all"
	case "2", "all except self", "All except self":
		channel.Settings.WhichChannelsToUse = "except self"
	case "3", "self", "Self":
		channel.Settings.WhichChannelsToUse = "self"
	case "4", "custom", "Custom":
		channel.Settings.WhichChannelsToUse = "custom"
		channel.Settings.CustomChannelsToUse = customChannels
	}

	conversationIDs.add(SayByID(channelID, "Gathering emotes and broadcaster information...").ID)

	go twitch.GetLiveStatuses(false)
	ok := twitch.GetEmoteController(false, channel)
	if !ok {
		DeleteDiscordChannel(channelName.Arguments[0])
		SayByIDAndDelete(channelID, "Could not retrieve "+channelName.Arguments[0]+"'s emotes...")
	}

	err := global.UpdateChannels("add", channel)
	if err == nil {
		twitch.Join(channelName.Arguments[0])
		SayByID(channelID, channelName.Arguments[0]+" added successfully.")
	} else {
		DeleteDiscordChannel(channelName.Arguments[0])
		SayByIDAndDelete(channelID, err.Error())
	}
}

func updateDirective(channelID string, messageID string) {
	var conversationIDs MessageIDs

	defer func() {
		conversationIDs.delete(channelID)
		dialogueChannel = nil
	}()

	dialogueChannel = make(chan Dialogue)
	conversationIDs.add(messageID)
	var channel *global.Directive

	conversationIDs.add(SayByID(channelID, "Which channel will you update?").ID)
	channelName := <-dialogueChannel
	conversationIDs.add(channelName.MessageID)
	if channelName.Arguments[0] == "cancel" {
		return
	}

	found := false
	for i, directive := range *&global.Directives {
		if strings.ToLower(channelName.Arguments[0]) == strings.ToLower(directive.ChannelName) {
			found = true
			channel = &global.Directives[i]
		}
	}

	if !found {
		conversationIDs.add(SayByID(channelID, "Not a tracked channel.").ID)
		return
	}

getWhatSettingToUpdate:
	conversationIDs.add(SayByID(channelID, spiel(channel)).ID)
	settingsToUpdate := <-dialogueChannel
	conversationIDs.add(settingsToUpdate.MessageID)
	if settingsToUpdate.Arguments[0] == "cancel" {
		return
	}
	if settingsToUpdate.Arguments[0] == "done" {
		goto decidedUpdates
	}
	for _, setting := range settingsToUpdate.Arguments {
		switch setting {
		default:
			conversationIDs.add(SayByID(channelID, "Not a proper answer.").ID)
			goto getWhatSettingToUpdate
		case "1":
			channel.Settings.IsCollectingMessages = !channel.Settings.IsCollectingMessages
			conversationIDs.add(SayByID(channelID, "Collecting messages: "+strconv.FormatBool(channel.Settings.IsCollectingMessages)).ID)
		case "2":
			channel.Settings.Reply.IsEnabled = !channel.Settings.Reply.IsEnabled
			conversationIDs.add(SayByID(channelID, "Reply: "+strconv.FormatBool(channel.Settings.Reply.IsEnabled)).ID)
		case "3":
			channel.Settings.Reply.IsAllowedWhenOnline = !channel.Settings.Reply.IsAllowedWhenOnline
			conversationIDs.add(SayByID(channelID, "Reply online: "+strconv.FormatBool(channel.Settings.Reply.IsAllowedWhenOnline)).ID)
		case "4":
			conversationIDs.add(SayByID(channelID, "New minutes to wait before replying online again?").ID)
			time := <-dialogueChannel
			conversationIDs.add(time.MessageID)
			timeParsed, success := parseTimeToWait(channelID, time.Arguments)
			if !success {
				DeleteDiscordChannel(channelName.Arguments[0])
				return
			}
			channel.Settings.Reply.OnlineTimeToWait = timeParsed
			conversationIDs.add(SayByID(channelID, "New minutes to wait before replying online again: "+time.Arguments[0]).ID)
		case "5":
			channel.Settings.Reply.IsAllowedWhenOffline = !channel.Settings.Reply.IsAllowedWhenOffline
			conversationIDs.add(SayByID(channelID, "Reply offline: "+strconv.FormatBool(channel.Settings.Reply.IsAllowedWhenOffline)).ID)
		case "6":
			conversationIDs.add(SayByID(channelID, "New minutes to wait before replying offline again?").ID)
			time := <-dialogueChannel
			conversationIDs.add(time.MessageID)
			timeParsed, success := parseTimeToWait(channelID, time.Arguments)
			if !success {
				DeleteDiscordChannel(channelName.Arguments[0])
				return
			}
			channel.Settings.Reply.OnlineTimeToWait = timeParsed
			conversationIDs.add(SayByID(channelID, "New minutes to wait before replying offline again: "+time.Arguments[0]).ID)
		case "7":
			channel.Settings.Participation.IsEnabled = !channel.Settings.Participation.IsEnabled
			conversationIDs.add(SayByID(channelID, "Participation: "+strconv.FormatBool(channel.Settings.Participation.IsEnabled)).ID)
		case "8":
			channel.Settings.Participation.IsAllowedWhenOnline = !channel.Settings.Participation.IsAllowedWhenOnline
			conversationIDs.add(SayByID(channelID, "Participation online: "+strconv.FormatBool(channel.Settings.Participation.IsAllowedWhenOnline)).ID)
		case "9":
			conversationIDs.add(SayByID(channelID, "New minutes to wait before participation online again?").ID)
			time := <-dialogueChannel
			conversationIDs.add(time.MessageID)
			timeParsed, success := parseTimeToWait(channelID, time.Arguments)
			if !success {
				DeleteDiscordChannel(channelName.Arguments[0])
				return
			}
			channel.Settings.Participation.OnlineTimeToWait = timeParsed
			conversationIDs.add(SayByID(channelID, "New minutes to wait before participation online again: "+time.Arguments[0]).ID)
		case "10":
			channel.Settings.Participation.IsAllowedWhenOffline = !channel.Settings.Participation.IsAllowedWhenOffline
			conversationIDs.add(SayByID(channelID, "Participation offline: "+strconv.FormatBool(channel.Settings.Participation.IsAllowedWhenOffline)).ID)
		case "11":
			conversationIDs.add(SayByID(channelID, "New minutes to wait before participation offline again?").ID)
			time := <-dialogueChannel
			conversationIDs.add(time.MessageID)
			timeParsed, success := parseTimeToWait(channelID, time.Arguments)
			if !success {
				DeleteDiscordChannel(channelName.Arguments[0])
				return
			}
			channel.Settings.Participation.OfflineTimeToWait = timeParsed
			conversationIDs.add(SayByID(channelID, "New minutes to wait before participation offline again: "+time.Arguments[0]).ID)
		case "12":
			conversationIDs.add(SayByID(channelID, "What chains will this channel use to post with?\n\nAll (1)     All except self (2)     Self (3)     Custom (4)\n\nIf custom, what are the custom channels to use?").ID)
			responseSettings := <-dialogueChannel
			conversationIDs.add(responseSettings.MessageID)
			if responseSettings.Arguments[0] == "cancel" {
				return
			}
			mode := responseSettings.Arguments[0]
			customChannels := responseSettings.Arguments[1:]
			switch mode {
			case "1", "all", "All":
				channel.Settings.WhichChannelsToUse = "all"
			case "2", "all except self", "All except self":
				channel.Settings.WhichChannelsToUse = "except self"
			case "3", "self", "Self":
				channel.Settings.WhichChannelsToUse = "self"
			case "4", "custom", "Custom":
				channel.Settings.WhichChannelsToUse = "custom"
				channel.Settings.CustomChannelsToUse = customChannels
			}
			conversationIDs.add(SayByID(channelID, "Participation offline: "+strings.Join(channel.Settings.CustomChannelsToUse, " ")).ID)
		}
	}
	goto getWhatSettingToUpdate

decidedUpdates:
	err := global.UpdateChannels("update", *channel)
	if err == nil {
		SayByID(channelID, strings.Title(channelName.Arguments[0])+" updated successfully.")
	} else {
		SayByIDAndDelete(channelID, err.Error())
	}
}

func removeDirective(channelID string, messageID string, args []string) {
	var conversationIDs MessageIDs

	defer func() {
		conversationIDs.delete(channelID)
		dialogueChannel = nil
	}()

	dialogueChannel = make(chan Dialogue)
	conversationIDs.add(messageID)

	if len(args) == 0 {
		conversationIDs.add(SayByID(channelID, "Enter channel to remove.").ID)
		channel := <-dialogueChannel
		conversationIDs.add(channel.MessageID)
		global.UpdateChannels("remove", global.Directive{ChannelName: channel.Arguments[0]})
		twitch.Depart(channel.Arguments[0])
	} else {
		global.UpdateChannels("remove", global.Directive{ChannelName: args[0]})
		twitch.Depart(args[0])
	}

	twitch.Depart(args[0])
}

func showDirectiveDetailed(channelID string, messageID string, args []string) {
	defer DeleteDiscordMessage(channelID, messageID)
	if len(args) < 1 {
		SayByIDAndDelete(channelID, "Specify channel!")
		return
	}
	for _, directive := range global.Directives {
		if directive.ChannelName == args[0] {
			file, err := json.MarshalIndent(directive, "", " ")
			if err != nil {
				SayByIDAndDelete(channelID, "There was an error!")
			}
			SayByIDAndDelete(channelID, string(file))
			return
		}
	}
	SayByIDAndDelete(channelID, "Channel "+args[0]+" not found in list.")
}

func showRegex(channelID string, messageID string) {
	defer DeleteDiscordMessage(channelID, messageID)
	SayByIDAndDelete(channelID, strings.Join(global.RegexList, ",\n"))
}

func addRegex(channelID string, messageID string, args []string) {
	defer DeleteDiscordMessage(channelID, messageID)

	if len(args) == 0 {
		go SayByIDAndDelete(channelID, "No regex provided.")
		return
	}

	for _, regexToAdd := range args {
		exists := false
		for _, regexExisting := range global.RegexList {
			if regexExisting == regexToAdd {
				go SayByIDAndDelete(channelID, regexToAdd+" is already added.")
				exists = true
			}
		}

		if !exists {
			global.RegexList = append(global.RegexList, regexToAdd)
		}
	}

	err := global.UpdateRegex()
	if err != nil {
		go SayByIDAndDelete(channelID, "Error:\n"+err.Error())
	} else {
		go SayByIDAndDelete(channelID, "Regex successfully updated.")
	}
}

func removeRegex(channelID string, messageID string, args []string) {
	defer DeleteDiscordMessage(channelID, messageID)

	if len(args) == 0 {
		go SayByIDAndDelete(channelID, "No regex provided.")
		return
	}

	for _, regexToRemove := range args {
		exists := false
		for i, regexExisting := range global.RegexList {
			if regexToRemove == regexExisting {
				global.RegexList = global.FastRemove(global.RegexList, i)
				exists = true
				break
			}
		}

		if !exists {
			go SayByIDAndDelete(channelID, regexToRemove+" is not on the list.")
		}
	}

	err := global.UpdateRegex()
	if err != nil {
		go SayByIDAndDelete(channelID, "Error:\n"+err.Error())
	} else {
		go SayByIDAndDelete(channelID, "Regex successfully updated.")
	}
}

func showBannedUsers(channelID string, messageID string) {
	defer DeleteDiscordMessage(channelID, messageID)
	SayByIDAndDelete(channelID, strings.Join(global.BannedUsers, ",\n"))
}

func addBannedUser(channelID string, messageID string, args []string) {
	defer DeleteDiscordMessage(channelID, messageID)

	if len(args) == 0 {
		go SayByIDAndDelete(channelID, "No users provided.")
		return
	}

	for _, userToAdd := range args {
		exists := false
		for _, userExisting := range global.BannedUsers {
			if userExisting == userToAdd {
				go SayByIDAndDelete(channelID, userToAdd+" is already added.")
				exists = true
			}
		}

		if !exists {
			global.BannedUsers = append(global.BannedUsers, userToAdd)
		}
	}

	err := global.SaveBannedUsers()
	if err != nil {
		go SayByIDAndDelete(channelID, "Error:\n"+err.Error())
	} else {
		go SayByIDAndDelete(channelID, "Banned users successfully updated.")
	}
}

func removeBannedUser(channelID string, messageID string, args []string) {
	defer DeleteDiscordMessage(channelID, messageID)

	if len(args) == 0 {
		go SayByIDAndDelete(channelID, "No regex provided.")
		return
	}

	for _, userToRemove := range args {
		exists := false
		for i, userExisting := range global.BannedUsers {
			if userToRemove == userExisting {
				global.BannedUsers = global.FastRemove(global.BannedUsers, i)
				exists = true
				break
			}
		}

		if !exists {
			go SayByIDAndDelete(channelID, userToRemove+" is not on the list.")
		}
	}

	err := global.SaveBannedUsers()
	if err != nil {
		go SayByIDAndDelete(channelID, "Error:\n"+err.Error())
	} else {
		go SayByIDAndDelete(channelID, "Banned users successfully updated.")
	}
}

func cleanse(channelID string, messageID string, args []string) {
	var conversationIDs MessageIDs

	defer func() {
		conversationIDs.delete(channelID)
		dialogueChannel = nil
	}()

	conversationIDs.add(messageID)
	conversationIDs.add(SayByID(channelID, "Cleansing chains of the word: "+args[0]).ID)
	cleansedNumber := markov.Cleanse(args[0])
	SayByID(channelID, "Cleansed a total of "+strconv.Itoa(cleansedNumber)+" entries matching ["+args[0]+"]")
}

func help(channelID string, messageID string) {
	defer DeleteDiscordMessage(channelID, messageID)
	SayByIDAndDelete(channelID, fmt.Sprintf("Commands:\n[%s]\n[%s]\n[%s]\n[%s]\n[%s]\n[%s]\n[%s]\n[%s]\n[%s]\n[%s]\n[%s]\n[%s]\n[%s]", "showchannels", "showchanneldetailed", "addchannel", "updatechannel", "removechannel", "showregex", "addregex", "removeregex", "showbannedusers", "addbanneduser", "removebanneduser", "cleanse", "help"))
}
