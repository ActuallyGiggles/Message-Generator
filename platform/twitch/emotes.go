package twitch

import (
	"Message-Generator/global"
	"Message-Generator/print"
	"sync"
)

var (
	Broadcasters   = make(map[string]Data)
	broadcastersMx sync.Mutex

	globalEmotesToUpdate              []global.Emote
	twitchChannelEmotesToUpdate       []global.Emote
	thirdPartyChannelEmotesToUpdate   []global.ThirdPartyEmotes
	thirdPartyChannelEmotesToUpdateMx sync.Mutex
)

func GetEmoteController(isInit bool, channel global.Directive) (ok bool) {
	broadcastersMx.Lock()
	thirdPartyChannelEmotesToUpdateMx.Lock()
	defer broadcastersMx.Unlock()
	defer thirdPartyChannelEmotesToUpdateMx.Unlock()
	if isInit {
		Broadcasters = make(map[string]Data)
	}

	if channel.ChannelName == "" {
		for _, directive := range global.Directives {
			routineBroadcastersUpdate(directive)
			if isInit {
				pb.UpdateTitle("Getting broadcasters info...")
				pb.Increment()
			}
		}

		if isInit {
			pb.UpdateTitle("Getting global emotes")
			getTwitchGlobalEmotes()
			pb.Increment()
			get7tvGlobalEmotes()
			pb.Increment()
			getBttvGlobalEmotes()
			pb.Increment()
			getFfzGlobalEmotes()
			pb.Increment()
		}

		for _, c := range Broadcasters {
			if isInit {
				pb.UpdateTitle("Getting channel emotes: " + c.Login + "...")
			}
			getTwitchChannelEmotes(c)
			if isInit {
				pb.Increment()
			}
			get7tvChannelEmotes(c)
			if isInit {
				pb.Increment()
			}
			getBttvChannelEmotes(c)
			if isInit {
				pb.Increment()
			}
			getFfzChannelEmotes(c)
			if isInit {
				pb.Increment()
			}
		}

		transferEmotes(isInit)
	} else {
		// Get Broadcaster Info
		data, err := GetBroadcasterInfo(channel.ChannelName)
		if err != nil {
			print.Error(err.Error())
			return false
		}
		// Add broadcaster
		Broadcasters[channel.ChannelName] = data

		// Get Twitch Channel Emotes
		err = getTwitchChannelEmotes(data)
		if err != nil {
			print.Error(err.Error())
		}
		// Add each twitch channel emote
		global.TwitchChannelEmotes = append(global.TwitchChannelEmotes, twitchChannelEmotesToUpdate...)
		twitchChannelEmotesToUpdate = nil

		thirdPartyChannelEmotesToUpdate = append(thirdPartyChannelEmotesToUpdate, global.ThirdPartyEmotes{Name: channel.ChannelName})

		// Get 7tv emotes
		err = get7tvChannelEmotes(data)
		if err != nil {
			print.Error(err.Error())
			return false
		}

		// Get BTTV emotes
		err = getBttvChannelEmotes(data)
		if err != nil {
			print.Error(err.Error())
			return false
		}

		// Get FFZ emotes
		err = getFfzChannelEmotes(data)
		if err != nil {
			print.Error(err.Error())
			return false
		}

		// Add each 7tv, BTTV, FFZ emote
		e := global.ThirdPartyEmotes{
			Name:   channel.ChannelName,
			Emotes: thirdPartyChannelEmotesToUpdate[0].Emotes,
		}
		global.ThirdPartyChannelEmotes = append(global.ThirdPartyChannelEmotes, e)
		thirdPartyChannelEmotesToUpdate = nil
	}

	return true
}

func transferEmotes(isInit bool) {
	global.EmotesMx.Lock()
	defer global.EmotesMx.Unlock()

	if isInit {
		transferGlobalEmotes()
	}

	transferTwitchChannelEmotes()
	transferThirdPartyEmotes()
}

func transferGlobalEmotes() {
	global.GlobalEmotes = nil
	global.GlobalEmotes = append(global.GlobalEmotes, globalEmotesToUpdate...)
	globalEmotesToUpdate = nil
}

func transferTwitchChannelEmotes() {
	global.TwitchChannelEmotes = nil
	global.TwitchChannelEmotes = append(global.TwitchChannelEmotes, twitchChannelEmotesToUpdate...)
	twitchChannelEmotesToUpdate = nil
}

func transferThirdPartyEmotes() {
	global.ThirdPartyChannelEmotes = nil
	global.ThirdPartyChannelEmotes = append(global.ThirdPartyChannelEmotes, thirdPartyChannelEmotesToUpdate...)
	thirdPartyChannelEmotesToUpdate = nil
}
