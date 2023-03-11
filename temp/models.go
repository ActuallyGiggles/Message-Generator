package temp

type List struct {
	Dates []struct {
		Year  string `json:"year"`
		Month string `json:"month"`
		Day   string `json:"day"`
	} `json:"availableLogs"`
}

type Log struct {
	Messages []struct {
		Channel  string `json:"channel"`
		Username string `json:"username"`
		Text     string `json:"text"`
	} `json:"messages"`
}
