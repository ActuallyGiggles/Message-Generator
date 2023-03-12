package temp

import "sync"

type Log struct {
	Messages []struct {
		Channel  string `json:"channel"`
		Username string `json:"username"`
		Text     string `json:"text"`
	} `json:"messages"`
}

type Streamer struct {
	Name     string
	Mutex    sync.Mutex
	DaysDone int
}

type Ready struct {
	Name   string
	Status string
}
