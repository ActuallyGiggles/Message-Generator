package main

import (
	"context"
	"markov-generator/api"
	"markov-generator/discord"
	"markov-generator/global"
	"markov-generator/handlers"
	"markov-generator/platform"
	"markov-generator/platform/twitch"
	"markov-generator/print"
	"markov-generator/stats"
	"markov-generator/twitter"
	"time"

	"os/signal"
	"syscall"

	"markov-generator/markov"

	"github.com/pkg/profile"
)

var debug = false

func main() {
	// Profiling
	defer profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()

	// Keep open
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	defer cancel()

	print.Page("Started")
	Start()

	go print.TerminalInput(cancel)

	<-ctx.Done()
	print.Page("Exiting")

	noted := false
	for {
		if markov.IsBusy() {
			if !noted {
				print.Info("Markov is busy.")
				noted = true
			}
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}

	print.Page("Exited")
	print.Info("Come back soon. T-T")
}

func Start() {
	// Make platform and discord channels
	incomingMessages := make(chan platform.Message)

	global.Start()
	go handlers.Incoming(incomingMessages)
	go api.HandleRequests()

	markov.Start(markov.StartInstructions{
		WriteInterval: 10,
		IntervalUnit:  "minutes",
		SeparationKey: " ",
		StartKey:      "b5G(n1$I!4g",
		EndKey:        "e1$D(n7",
		Debug:         false,
		ShouldZip:     true,
		ShouldDefluff: true,
	})

	go twitter.Start()

	go discord.Start()

	twitch.GatherEmotes(debug)

	go twitch.Start(incomingMessages)

	stats.Start()

	print.Page("Twitch Message Generator")
	print.Started("Program Started at " + time.Now().Format(time.RFC822))
}
