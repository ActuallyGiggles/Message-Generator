package main

import (
	"Message-Generator/handlers"
	"Message-Generator/markov"
	"Message-Generator/platform"
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pterm/pterm"
)

var channelsToWrite int
var channelsWritten int
var timesWritten uint64

func doIt() {
	spinner, err := pterm.DefaultSpinner.Start(fmt.Sprintf("[%v] Started Chaining Process", time.Now().Format(time.Stamp)))
	if err != nil {
		panic(err)
	}
	var wgMain sync.WaitGroup
	for _, streamer := range getLogFolders() {
		wgMain.Add(1)
		channelsToWrite++
		go processStreamerLogs(streamer, spinner, &wgMain)
	}
	wgMain.Wait()
	spinner.Success("[%v] Finished Chaining Process", time.Now().Format(time.Stamp))
}

func processStreamerLogs(streamer string, spinner *pterm.SpinnerPrinter, wgMain *sync.WaitGroup) {
	months, err := os.ReadDir("./collected-logs/" + streamer + "/")
	if err != nil {
		panic(err)
	}

	var linesPassed int

	for _, month := range months {
		file, err := os.Open("./collected-logs/" + streamer + "/" + month.Name())
		if err != nil {
			panic(err)
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if scanner.Text() == "" {
				continue
			}

			split := strings.Split(scanner.Text(), " ")

			if handlers.TempIncoming(platform.Message{
				Platform:    "twitch",
				ChannelName: streamer,
				AuthorName:  strings.ReplaceAll(split[3], ":", ""),
				Content:     strings.Join(split[4:], " "),
			}) {
				linesPassed++
			}

			if linesPassed < 10000 {
				continue
			}

		loop:
			time.Sleep(time.Second)
			if markov.ChainIntake(streamer) == -1 {
				goto loop
			}
			markov.TempTriggerWrite(streamer)
			timesWritten++
			spinner.UpdateText(fmt.Sprintf("[%v] Chaining... Channels Chained: %03d/%03d | Times Written: %d", time.Now().Format(time.Stamp), channelsWritten, channelsToWrite, timesWritten))
			linesPassed = 0
		}

		if err := scanner.Err(); err != nil {
			panic(err)
		}

		file.Close()
	}

loop2:
	time.Sleep(time.Second)
	if markov.ChainIntake(streamer) == -1 {
		goto loop2
	}

	markov.TempTriggerWrite(streamer)
	timesWritten++
	spinner.UpdateText(fmt.Sprintf("[%v] Chaining... Channels Chained: %03d/%03d | Times Written: %d", time.Now().Format(time.Stamp), channelsWritten, channelsToWrite, timesWritten))
	wgMain.Done()
	channelsWritten++
	pterm.Success.Printf("[%v] Chained %s\n", time.Now().Format(time.Stamp), streamer)
}

func getLogFolders() (slice []string) {
	files, err := os.ReadDir("./collected-logs/")
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		slice = append(slice, file.Name())
	}

	return
}
