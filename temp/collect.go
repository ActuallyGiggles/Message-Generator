package temp

import (
	"Message-Generator/global"
	"Message-Generator/markov"
	"Message-Generator/platform"
	"Message-Generator/print"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/pterm/pterm"
)

var streamers = make(map[string]*Streamer)
var streamerMx sync.Mutex
var totalChannels int
var doneChannels int
var daysDone uint
var logsAccess sync.Mutex

func Start(c chan platform.Message) {
	spinner := print.Spinner("Starting Transcription...")

	var wg sync.WaitGroup
	for _, directive := range global.Directives {
		wg.Add(1)
		go transcribeChannel(c, directive, spinner, &wg)
	}

	wg.Wait()
	spinner.Success(fmt.Sprintf("All %d Logs Completed!", totalChannels))
}

func transcribeChannel(c chan platform.Message, directive global.Directive, spinner *pterm.SpinnerPrinter, wg *sync.WaitGroup) {
	newDate := time.Now()
	firstDay := true

collectAnotherDaysLogs:
	log, issue := collectLogForDay(directive.ChannelName, newDate.Year(), newDate.Month(), newDate.Day())
	if issue != "" {
		if firstDay {
			return
		}

		doneChannels++
		print.Success("Finished Writing Logs for " + directive.ChannelName + " at day" + newDate.Format(time.DateOnly))
		wg.Done()
		spinner.UpdateText(fmt.Sprintf("Transcribing Channels... %d/%d (days written: %d)", doneChannels, totalChannels, daysDone))
		return
	} else {
		if firstDay {
			firstDay = false
			totalChannels++
		}
	}

	for _, message := range log.Messages {
		c <- platform.Message{
			ChannelName: message.Channel,
			AuthorName:  message.Username,
			Content:     message.Text,
		}
	}

	markov.TempTriggerWrite(directive.ChannelName)

	daysDone++
	spinner.UpdateText(fmt.Sprintf("Transcribing Channels... %d/%d (days written: %d)", doneChannels, totalChannels, daysDone))
	newDate = newDate.AddDate(0, 0, -1)
	goto collectAnotherDaysLogs
}

func collectLogForDay(channelName string, year int, month time.Month, day int) (log Log, issue string) {
	logsAccess.Lock()

	url := fmt.Sprintf("https://logs.ivr.fi/channel/%s/%s/%s/%s?json=", channelName, strconv.Itoa(year), month.String(), strconv.Itoa(day))

	var jsonStr = []byte(`{"content-type":"application/json"}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", "Bearer "+global.TwitchOAuth)
	req.Header.Set("Client-Id", global.TwitchClientID)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		panic(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	go allowLogsAccess()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, &log); err != nil {
		return log, string(body)
	}

	return log, ""
}

func saveWrittenChannelsAndDatesAsJson(obj map[string][]string) {
	jsonObj, err := json.MarshalIndent(obj, "", " ")
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("./temp/saved_logs.json", jsonObj, 0644)
	if err != nil {
		panic(err)
	}
}

func allowLogsAccess() {
	time.Sleep(time.Second)
	logsAccess.Unlock()
}
