package temp

import (
	"Message-Generator/global"
	"Message-Generator/markov"
	"Message-Generator/platform"
	"Message-Generator/print"
	"fmt"
	"sync"
	"time"

	"github.com/pterm/pterm"
)

var spinner *pterm.SpinnerPrinter
var streamers = make(map[string]*sync.Mutex)
var streamersMx sync.Mutex
var statusOfGoRoutines = make(chan Status)a
var readyChannels []string
var totalChannels int
var doneChannels int
var logsAccess sync.Mutex

func Start(c chan platform.Message) {
	spinner = print.Spinner("Starting Transcription...")

	channels := returnListOfAvailableStreamers().Channels
	for _, channel := range channels {
		streamers[channel.Name] = &sync.Mutex{}
	}

	go StatusManager()

	var wg sync.WaitGroup
	for _, streamer := range channels {
		for _, directive := range global.Directives {
			if streamer.Name == directive.ChannelName {
				wg.Add(1)
				totalChannels++
				go transcribe(c, directive, &wg)
			}
		}
	}

	wg.Wait()
	spinner.Success(fmt.Sprintf("All %d Logs Completed!", totalChannels))
}

func transcribe(c chan platform.Message, directive global.Directive, wg *sync.WaitGroup) {
	defer wg.Done()
	newDate := time.Now()

	for {
		if newDate.Year() == 2020 {
			doneChannels++
			print.Success("Finished Writing Logs for " + directive.ChannelName + " up to day " + newDate.Format(time.RFC3339))
			statusOfGoRoutines <- Status{Name: directive.ChannelName, Status: "complete"}
			return
		}

		theMutex := streamers[directive.ChannelName]
		theMutex.Lock()

		log, _ := collectLogForDay(directive.ChannelName, newDate.Year(), newDate.Month(), newDate.Day())
		for _, message := range log.Messages {
			c <- platform.Message{
				ChannelName: message.Channel,
				AuthorName:  message.Username,
				Content:     message.Text,
			}
		}

		statusOfGoRoutines <- Status{Name: directive.ChannelName, Status: "ready"}
		newDate = newDate.AddDate(0, 0, -1)
	}
}

func StatusManager() {
	for status := range statusOfGoRoutines {
		go func(status Status) {
			if status.Status == "ready" {
				readyChannels = append(readyChannels, status.Name)
			}

			if status.Status == "complete" {
				totalChannels--
			}

			print.InfoNoTime(fmt.Sprintf("%s %s (%d/%d)", status.Name, status.Status, len(readyChannels), totalChannels))
			spinner.UpdateText(fmt.Sprintf("Transcribing Channels... %d/%d (channels ready: %d/%d)", doneChannels, totalChannels, len(readyChannels), totalChannels))

			if len(readyChannels) >= totalChannels {
				print.InfoNoTime("writing")
				markov.TempTriggerWrite()
				for _, channel := range readyChannels {
					for s, mutex := range streamers {
						if s == channel {
							mutex.Unlock()
						}
					}
				}
				readyChannels = nil
				print.InfoNoTime("done writing")
			}
		}(status)
	}
}

func allowLogsAccess() {
	time.Sleep(time.Second)
	logsAccess.Unlock()
}
