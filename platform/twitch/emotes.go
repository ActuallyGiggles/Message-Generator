package twitch

import (
	"Message-Generator/global"
	"Message-Generator/stats"
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
	Broadcasters = make(map[string]Data)

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
			stats.Log(err.Error())
			return false
		}
		Broadcasters[channel.ChannelName] = data // Add broadcaster

		// Get Twitch Channel Emotes
		err = getTwitchChannelEmotes(data)
		if err != nil {
			stats.Log(err.Error())
		}
		global.TwitchChannelEmotes = append(global.TwitchChannelEmotes, twitchChannelEmotesToUpdate...) // Add each twitch channel emote
		twitchChannelEmotesToUpdate = nil

		thirdPartyChannelEmotesToUpdate = append(thirdPartyChannelEmotesToUpdate, global.ThirdPartyEmotes{Name: channel.ChannelName})

		// Get 7tv emotes
		err = get7tvChannelEmotes(data)
		if err != nil {
			stats.Log(err.Error())
			return false
		}

		// Get BTTV emotes
		err = getBttvChannelEmotes(data)
		if err != nil {
			stats.Log(err.Error())
			return false
		}

		// Get FFZ emotes
		err = getFfzChannelEmotes(data)
		if err != nil {
			stats.Log(err.Error())
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
	// fmt.Printf("%d Global emotes\n", len(global.GlobalEmotes))
}

func transferTwitchChannelEmotes() {
	global.TwitchChannelEmotes = nil
	global.TwitchChannelEmotes = append(global.TwitchChannelEmotes, twitchChannelEmotesToUpdate...)
	twitchChannelEmotesToUpdate = nil
	// fmt.Printf("%d Twitch Channel emotes\n", len(global.TwitchChannelEmotes))
}

func transferThirdPartyEmotes() {
	global.ThirdPartyChannelEmotes = nil
	global.ThirdPartyChannelEmotes = append(global.ThirdPartyChannelEmotes, thirdPartyChannelEmotesToUpdate...)
	thirdPartyChannelEmotesToUpdate = nil
	// fmt.Printf("%d Third Party emotes\n", func() (total int) {
	// 	for _, c := range global.ThirdPartyChannelEmotes {
	// 		total += len(c.Emotes)
	// 	}
	// 	return
	// }())
}
