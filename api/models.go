package api

import (
	"markov-generator/platform/twitch"
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
