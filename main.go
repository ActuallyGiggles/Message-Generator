package main

import (
	"Message-Generator/api"
	"Message-Generator/discord"
	"Message-Generator/global"
	"Message-Generator/handlers"
	"Message-Generator/platform"
	"Message-Generator/platform/twitch"
	"Message-Generator/print"
	"Message-Generator/stats"
	"Message-Generator/twitter"
	"context"
	"time"

	"os/signal"
	"syscall"

	"Message-Generator/markov"

	"github.com/pkg/profile"
)

var debug = true

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

	go twitter.Start()
	go discord.Start()

	markov.Start(markov.StartInstructions{
		SeparationKey:       " ",
		StartKey:            "b5G(n1$I!4g",
		EndKey:              "e1$D(n7",
		Debug:               true,
		ShouldZip:           true,
		ShouldDefluff:       true,
		DefluffTriggerValue: 50,
	})

	twitch.GatherEmotes(debug)
	go twitch.Start(debug, incomingMessages)

	stats.Start()

	print.Page("Twitch Message Generator")
	print.Started("Program Started at " + time.Now().Format(time.RFC822))
}
