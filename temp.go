package main

import (
	"Message-Generator/markov"
	"Message-Generator/platform"
	"bufio"
	"os"
	"strings"
)

var streamerMonthsAlreadySeen = make(map[string][]string)

func doIt(c chan platform.Message) {

	// Get all the streamers folders that exist and put it into the map streamerMonthsAlreadySeen
	getLogFolders()

	for streamer, monthsSeen := range streamerMonthsAlreadySeen {
		months, err := os.ReadDir("./collected-logs/" + streamer + "/")
		if err != nil {
			panic(err)
		}

	monthForLoop:
		for _, month := range months {
			// If month was already processed, skip it
			for _, monthSeen := range monthsSeen {
				if monthSeen == month.Name() {
					continue monthForLoop
				}
			}

			f, err := os.Open("./collected-logs/" + streamer + "/" + month.Name())
			if err != nil {
				panic(err)
			}

			fileScanner := bufio.NewScanner(f)
			fileScanner.Split(bufio.ScanLines)
			for fileScanner.Scan() {
				split := strings.Split(fileScanner.Text(), " ")
				author := strings.ReplaceAll(split[3], ":", "")
				content := strings.Join(split[4:], " ")

				message := platform.Message{
					Platform:    "twitch",
					ChannelName: streamer,
					AuthorName:  author,
					Content:     content,
				}

				c <- message
			}

			markov.TempTriggerWriteTicker()

			f.Close()
			streamerMonthsAlreadySeen[streamer] = append(streamerMonthsAlreadySeen[streamer], month.Name())

		}
	}
}

func getLogFolders() {
	files, err := os.ReadDir("./collected-logs/")
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		streamerMonthsAlreadySeen[file.Name()] = nil
	}
}
