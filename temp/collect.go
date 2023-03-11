package temp

import (
	"Message-Generator/global"
	"Message-Generator/markov"
	"Message-Generator/platform"
	"Message-Generator/print"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func Start(c chan platform.Message) {
	var channelAndDatesDone = make(map[string][]string)

	spinner := print.Spinner("Writing logs...")

	for i, directive := range global.Directives {
		spinner.UpdateText(fmt.Sprintf("%d Logs Completed for %s - Channels Finished: %d/%d", 0, directive.ChannelName, i, len(global.Directives)))

		list, success := collectListPerChannel(directive.ChannelName)
		if !success {
			continue
		}

		channelAndDatesDone[directive.ChannelName] = []string{}

		for j, date := range list.Dates {
			log, success, issue := collectLogForDay(directive.ChannelName, date.Year, date.Month, date.Day)
			if !success {
				panic(errors.New(issue))
			}

			for _, message := range log.Messages {
				c <- platform.Message{
					ChannelName: message.Channel,
					AuthorName:  message.Username,
					Content:     message.Text,
				}
			}

			markov.TempTriggerWrite()

			channelAndDatesDone[directive.ChannelName] = append(channelAndDatesDone[directive.ChannelName], date.Year+"/"+date.Month+"/"+date.Day)
			saveWrittenChannelsAndDatesAsJson(channelAndDatesDone)

			spinner.UpdateText(fmt.Sprintf("%d Logs Completed for %s - Channels Finished: %d/%d", j+1, directive.ChannelName, i, len(global.Directives)))
		}

		print.Success("Finished Writing Logs for " + directive.ChannelName)
	}

	spinner.Success(fmt.Sprintf("All %d Logs Completed!", len(global.Directives)))
}

func collectListPerChannel(channelName string) (list List, success bool) {
	url := "https://logs.ivr.fi/list?channel=" + channelName

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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, &list); err != nil {
		return list, false
	}

	return list, true
}

func collectLogForDay(channelName, year, month, day string) (log Log, success bool, issue string) {
	url := fmt.Sprintf("https://logs.ivr.fi/channel/%s/%s/%s/%s?json=", channelName, year, month, day)

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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, &log); err != nil {
		return log, false, string(body)
	}

	return log, true, ""
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
