package twitch

import (
	"Message-Generator/global"
	"Message-Generator/print"
	"sync"
	"time"

	"github.com/pterm/pterm"
)

var (
	IsLive   = make(map[string]bool)
	IsLiveMx sync.Mutex
	pb       *pterm.ProgressbarPrinter
)

func GatherEmotes(debug bool) {
	pb = print.ProgressBar("Collecting Twitch API information...", 4+len(global.Directives)*6)
	GetLiveStatuses(true)

	if debug {
		return
	}

	GetEmoteController(true, global.Directive{})
	pb.Stop()

	go updateLiveStatuses()
	go refreshEmotes()
}

func updateLiveStatuses() {
	for range time.Tick(2 * time.Minute) {
		GetLiveStatuses(false)
	}
}

func refreshEmotes() {
	for range time.Tick(30 * time.Minute) {
		GetEmoteController(false, global.Directive{})
	}
}

func IsChannelLive(channel string) bool {
	IsLiveMx.Lock()
	defer IsLiveMx.Unlock()
	return IsLive[channel]
}
