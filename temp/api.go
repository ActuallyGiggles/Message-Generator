package temp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func returnListOfAvailableStreamers() (listOfStreamers AvailableStreamers) {
	logsAccess.Lock()

	url := "https://logs.ivr.fi/channels"

	var jsonStr = []byte(`{"content-type":"application/json"}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
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
	if err := json.Unmarshal(body, &listOfStreamers); err != nil {
		panic(err)
	}

	return listOfStreamers
}

func collectLogForDay(channelName string, year int, month time.Month, day int) (log Log, issue string) {
	logsAccess.Lock()

	url := fmt.Sprintf("https://logs.ivr.fi/channel/%s/%s/%s/%s?json=", channelName, strconv.Itoa(year), month.String(), strconv.Itoa(day))

	var jsonStr = []byte(`{"content-type":"application/json"}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
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
