package temp

type Log struct {
	Messages []struct {
		Channel  string `json:"channel"`
		Username string `json:"username"`
		Text     string `json:"text"`
	} `json:"messages"`
}

type AvailableStreamers struct {
	Channels []struct {
		Name   string `json:"name"`
		UserID string `json:"userID"`
	} `json:"channels"`
}

type Status struct {
	Name   string
	Status string
}
