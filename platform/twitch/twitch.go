package twitch

import (
	"Message-Generator/global"
	"Message-Generator/platform"
	"Message-Generator/print"
	"time"

	"github.com/gempir/go-twitch-irc/v3"
)

var client *twitch.Client

var totalM int

// Start creates a twitch client and connects it.
func Start(incoming chan platform.Message) {
startOver:
	// Make unexported client use the address for the initialized client
	client = &twitch.Client{}
	client = twitch.NewClient(global.BotName, "oauth:"+global.TwitchOAuth)

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		m := platform.Message{
			Platform:    "twitch",
			ChannelName: message.Channel,
			ChannelID:   message.ID,
			AuthorName:  message.User.Name,
			AuthorID:    message.User.ID,
			Content:     message.Message,
		}

		incoming <- m
	})

	for _, directive := range global.Directives {
		client.Join(directive.ChannelName)
	}

	err := client.Connect()
	if err != nil {
		time.Sleep(10 * time.Second)
		print.Error("started over:\n" + err.Error())
		goto startOver
	}
}

// Say sends a message to a specific twitch chatroom.
func Say(channel string, message string) {
	client.Say(channel, message)
}

// Join joins a twitch chatroom.
func Join(channel string) {
	client.Join(channel)
}

// Depart departs a twitch chatroom.
func Depart(channel string) {
	client.Depart(channel)
}
