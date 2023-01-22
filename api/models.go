package api

import (
	"Twitch-Message-Generator/platform/twitch"
)

type APIResponse struct {
	MarkovSentence string `json:"markov_sentence"`
	Error          string `json:"error"`
}

type DataSend struct {
	ChannelsUsed []twitch.Data
	ChannelsLive []ChannelsLive
}

type ChannelsLive struct {
	Name string
	Live bool
}
