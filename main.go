package main

import (
	webhook "Adam/discord-twoup/Discord"
	"Adam/discord-twoup/Handlers"
	"Adam/discord-twoup/MatchFinder"
	"context"
	"fmt"
	"log"
)

const url = "https://canary.discord.com/api/webhooks/1267996258203865201/ZenhYVAIrjy3YLLkqtAjSIEGJ3W6UW2Qn627G41IrcxiaAZrSTGjegYR9zDbwfUkmD4v"

func main() {

	var MatchHandler = &Handlers.MatchHandler{
		MatchNotifier: &Handlers.DiscordMatchNotifier{
			Ping:  Handlers.DiscordPingType{Type: "User", IDs: []string{"261512678269255681"}},
			Url:   url,
			Color: "1207889",
		},
	}

	qt := MatchFinder.TwoUp
	qR, err := MatchFinder.Find(qt)
	if err != nil {
		err := webhook.SendMessage(url, createPlainMessage(err.Error()))
		if err != nil {
			errStr := fmt.Sprintf("Error logging error: %v\n", err)
			log.Println(errStr)
			return
		}
	}
	err = MatchHandler.MatchNotifier.NotifyUser(context.Background(), &qR)
	if err != nil {
		fmt.Println("Failed Webhook:", err)
		err = webhook.SendMessage(url, logError(err))
		if err != nil {
			errStr := fmt.Sprintf("Error logging error: %v\n", err)
			log.Println(errStr)
			return
		}
	}
}

func createPlainMessage(content string) webhook.Message {
	message := webhook.Message{
		Content: &content,
	}
	return message
}

func logError(err error) webhook.Message {
	return createPlainMessage(err.Error())
}
