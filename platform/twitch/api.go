package twitch

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"markov-generator/global"
	"markov-generator/stats"
	"net/http"
)

func routineBroadcastersUpdate(directive global.Directive) (err error) {
	data, err := GetBroadcasterInfo(directive.ChannelName)
	if err == nil {
		Broadcasters[directive.ChannelName] = data

		e := global.ThirdPartyEmotes{
			Name: directive.ChannelName,
		}

		thirdPartyChannelEmotesToUpdate = append(thirdPartyChannelEmotesToUpdate, e)
	} else {
		return err
	}

	return nil
}

func GetBroadcasterInfo(channelName string) (data Data, err error) {
	url := "https://api.twitch.tv/helix/users?login=" + channelName

	d := Data{}
	var jsonStr = []byte(`{"content-type":"application/json"}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", "Bearer "+global.TwitchOAuth)
	req.Header.Set("Client-Id", global.TwitchClientID)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		stats.Log("GetBroadcasterID failed\n", err.Error())
		return d, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		stats.Log("GetBroadcasterID failed\n", err.Error())
		return d, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		log.Println("GetBroadcasterID failed (StatusCode not 200) for", "'"+channelName+"'", "\n"+string(body))
	}
	broadcaster := Broadcaster[Data]{}
	if err := json.Unmarshal(body, &broadcaster); err != nil {
		stats.Log("GetBroadcasterID failed\n", err.Error())
		return d, err
	}
	for _, v := range broadcaster.Data {
		d.ID = v.ID
		d.Login = v.Login
		d.DisplayName = v.DisplayName
		d.Type = v.Type
		d.BroadcasterType = v.BroadcasterType
		d.Description = v.Description
		d.ProfileImageUrl = v.ProfileImageUrl
		d.OfflineImageUrl = v.OfflineImageUrl
		d.ViewCount = v.ViewCount
		d.Email = v.Email
		d.CreatedAt = v.CreatedAt
	}
	return d, nil
}

func getTwitchGlobalEmotes() {
	url := "https://api.twitch.tv/helix/chat/emotes/global"

	var jsonStr = []byte(`{"":""}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", "Bearer "+global.TwitchOAuth)
	req.Header.Set("Client-Id", global.TwitchClientID)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		panic(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	emotes := TwitchEmoteAPIResponse[TwitchGlobalEmote]{}
	if err := json.Unmarshal(body, &emotes); err != nil {
		panic(err)
	}

	for _, emote := range emotes.Data {
		e := global.Emote{
			Name: emote.Name,
			Url:  emote.Images["url_4x"],
		}

		globalEmotesToUpdate = append(globalEmotesToUpdate, e)
	}
}

func getTwitchChannelEmotes(c Data) (err error) {
	url := "https://api.twitch.tv/helix/chat/emotes?broadcaster_id=" + c.ID

	var jsonStr = []byte(`{"":""}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", "Bearer "+global.TwitchOAuth)
	req.Header.Set("Client-Id", global.TwitchClientID)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return errors.New("GetTwitchChannelEmotes failed (StatusCode not 200) for " + c.Login + "\n" + string(body))
	}
	emotes := TwitchEmoteAPIResponse[TwitchChannelEmote]{}
	if err := json.Unmarshal(body, &emotes); err != nil {
		return err
	}

	for _, emote := range emotes.Data {
		e := global.Emote{
			Name: emote.Name,
			Url:  emote.Images["url_4x"],
		}

		twitchChannelEmotesToUpdate = append(twitchChannelEmotesToUpdate, e)
	}

	return nil
}

func get7tvGlobalEmotes() {
	url := "https://api.7tv.app/v2/emotes/global"

	var jsonStr = []byte(`{"":""}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	emotes := []SevenTVEmote{}
	if err := json.Unmarshal(body, &emotes); err != nil {
		log.Fatal(err)
	}

	for _, emote := range emotes {
		e := global.Emote{
			Name: emote.Name,
			Url:  emote.Urls[3][1],
		}

		globalEmotesToUpdate = append(globalEmotesToUpdate, e)
	}
}

func get7tvChannelEmotes(c Data) (err error) {
	url := "https://api.7tv.app/v2/users/" + c.Login + "/emotes"

	var jsonStr = []byte(`{"":""}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		// User not registered for 7tv
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	emotes := []SevenTVEmote{}
	if err := json.Unmarshal(body, &emotes); err != nil {
		return err
	}

	for i, channel := range thirdPartyChannelEmotesToUpdate {
		if channel.Name == c.Login {
			for _, emote := range emotes {
				thirdPartyChannelEmotesToUpdate[i].Emotes = append(thirdPartyChannelEmotesToUpdate[i].Emotes, global.Emote{
					Name: emote.Name,
					Url:  emote.Urls[3][1],
				})
			}
		}
	}

	return nil
}

func getBttvGlobalEmotes() {
	url := "https://api.betterttv.net/3/cached/emotes/global"

	var jsonStr = []byte(`{"":""}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		panic(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	emotes := []BttvEmote{}
	if err := json.Unmarshal(body, &emotes); err != nil {
		panic(err)
	}

	for _, emote := range emotes {
		e := global.Emote{
			Name: emote.Name,
			Url:  "https://cdn.betterttv.net/emote/" + emote.ID + "/3x.png",
		}

		globalEmotesToUpdate = append(globalEmotesToUpdate, e)
	}
}

func getBttvChannelEmotes(c Data) (err error) {
	url := "https://api.betterttv.net/3/cached/users/twitch/" + c.ID

	var jsonStr = []byte(`{"":""}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		// User not registered for 7tv
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	emotes := BttvChannelEmotes[BttvEmote]{}
	if err := json.Unmarshal(body, &emotes); err != nil {
		return err
	}

	for i, channel := range thirdPartyChannelEmotesToUpdate {
		if channel.Name == c.Login {
			for _, emote := range emotes.ChannelEmotes {
				thirdPartyChannelEmotesToUpdate[i].Emotes = append(thirdPartyChannelEmotesToUpdate[i].Emotes, global.Emote{
					Name: emote.Name,
					Url:  "https://cdn.betterttv.net/emote/" + emote.ID + "/3x.png",
				})
			}
			for _, emote := range emotes.SharedEmotes {
				thirdPartyChannelEmotesToUpdate[i].Emotes = append(thirdPartyChannelEmotesToUpdate[i].Emotes, global.Emote{
					Name: emote.Name,
					Url:  "https://cdn.betterttv.net/emote/" + emote.ID + "/3x.png",
				})
			}
		}
	}

	return nil
}

func getFfzGlobalEmotes() {
	url := "https://api.frankerfacez.com/v1/set/global"

	var jsonStr = []byte(`{"":""}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		panic(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	set := FfzSets{}
	if err := json.Unmarshal(body, &set); err != nil {
		panic(err)
	}

	for _, emotes := range set.Sets {
		for _, emote := range emotes.Emoticons {
			e := global.Emote{
				Name: emote.Name,
			}
			for size, url := range emote.Urls {
				switch size {
				case "4":
					e.Url = "https:" + url
					break
				case "2":
					e.Url = "https:" + url
					break
				case "1":
					e.Url = "https:" + url
					break
				}
			}
			globalEmotesToUpdate = append(globalEmotesToUpdate, e)
		}
	}
}

func getFfzChannelEmotes(c Data) (err error) {
	url := "https://api.frankerfacez.com/v1/room/id/" + c.ID

	var jsonStr = []byte(`{"":""}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		// User not registered for 7tv
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	set := FfzSets{}
	if err := json.Unmarshal(body, &set); err != nil {
		return err
	}

	for i, channel := range thirdPartyChannelEmotesToUpdate {
		if channel.Name == c.Login {
			for _, emotes := range set.Sets {
				for _, emote := range emotes.Emoticons {
					e := global.Emote{
						Name: emote.Name,
					}
					for size, url := range emote.Urls {
						switch size {
						case "4":
							e.Url = "https:" + url
							break
						case "2":
							e.Url = "https:" + url
							break
						case "1":
							e.Url = "https:" + url
							break
						}
					}
					thirdPartyChannelEmotesToUpdate[i].Emotes = append(thirdPartyChannelEmotesToUpdate[i].Emotes, e)
				}
			}
		}
	}

	return nil
}

func GetLiveStatus(channelName string) (live bool) {
	url := "https://api.twitch.tv/helix/streams?user_login=" + channelName
	var jsonStr = []byte(`{"":""}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", "Bearer "+global.TwitchOAuth)
	req.Header.Set("Client-Id", global.TwitchClientID)
	if err != nil {
		stats.Log(err.Error())
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		stats.Log(err.Error())
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var stream StreamStatusData
	if err := json.Unmarshal(body, &stream); err != nil {
		stats.Log(err.Error())
	}
	if len(stream.Data) == 0 {
		return false
	} else {
		return true
	}
}

func GetLiveStatuses(isInit bool) {
	IsLiveMx.Lock()
	defer IsLiveMx.Unlock()
	for _, channel := range global.Directives {
		r := GetLiveStatus(channel.ChannelName)
		IsLive[channel.ChannelName] = r
		if isInit {
			pb.UpdateTitle("Getting live statuses...")
			pb.Increment()
		}
	}
}
