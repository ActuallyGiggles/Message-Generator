package twitter

import (
	"Twitch-Message-Generator/global"
	"Twitch-Message-Generator/platform/twitch"
	"Twitch-Message-Generator/stats"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/tweet/managetweet"
	"github.com/michimani/gotwi/tweet/managetweet/types"
)

var client *gotwi.Client
var potentialTweets = make(map[string][]string)
var potentialTweetsMx sync.Mutex

func Start() {
	in := &gotwi.NewClientInput{
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		OAuthToken:           global.TwitterAccessToken,
		OAuthTokenSecret:     global.TwitterAccessTokenSecret,
	}

	c, err := gotwi.NewClient(in)
	client = c
	if err != nil {
		panic(fmt.Sprintf("Twitter not started.\n %e", err))
	}

	go pickTweet()
}

func SendTweet(channel string, message string) {
	for _, d := range twitch.Broadcasters {
		if d.Login == channel {
			channel = d.DisplayName
		}
	}

	message = fmt.Sprintf("%s\n\n#%sChatSays\n#ShitTwitchChatSays", message, channel)

	//log.Println(fmt.Sprintf("Tweet: \n\tChannel: %s \n\tMessage: %s", channel, strings.ReplaceAll(message, "ChatSays \n", "ChatSays ")))

	p := &types.CreateInput{
		Text: gotwi.String(message),
	}

	_, err := managetweet.Create(context.Background(), client, p)
	if err != nil {
		stats.Log(err.Error())
		return
	}
}

func AddMessageToPotentialTweets(channel string, message string) {
	// Add to map
	if len(message) > 227 {
		return
	}
	potentialTweetsMx.Lock()
	defer potentialTweetsMx.Unlock()
	potentialTweets[channel] = append(potentialTweets[channel], message)
}

func pickTweet() {
	// Create ticker to repeat tweet picking
	for range time.Tick(1 * time.Hour) {
		channel, message, empty := pickRandomFromMap()
		if empty {
			stats.Log("Empty map.")
		} else {
			SendTweet(channel, message)
		}
		potentialTweetsMx.Lock()
		potentialTweets = make(map[string][]string)
		potentialTweetsMx.Unlock()
	}
}

func pickRandomFromMap() (channel, message string, empty bool) {
	if len(potentialTweets) == 0 {
		return "", "", true
	}

	// Get slice of channels in map
	var channels []string
	for channel := range potentialTweets {
		channels = append(channels, channel)
	}

	// Get random channel
	channel = global.PickRandomFromSlice(channels)

	// Get random message from channel
	message = global.PickRandomFromSlice(potentialTweets[channel])

	return channel, message, false
}
