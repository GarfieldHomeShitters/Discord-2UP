package main

import (
	"Adam/discord-twoup/Handlers"
	"Adam/discord-twoup/MatchFinder"
	"context"
	"fmt"
	"log"
)

const url = "https://canary.discord.com/api/webhooks/1267996258203865201/ZenhYVAIrjy3YLLkqtAjSIEGJ3W6UW2Qn627G41IrcxiaAZrSTGjegYR9zDbwfUkmD4v"

type NotificationHandler struct {
	Handlers.MatchNotifier
	Handlers.ErrorHandler
}

func CreateNotificationHandler() *NotificationHandler {
	UserIDs := []string{"261512678269255681"}
	ping := Handlers.NewPingType(nil, UserIDs, nil)
	MatchHandler := Handlers.NewDiscordMatchNotifier(url, *ping, "1207889")
	ErrorHandler := Handlers.NewDiscordErrorHandler(url, "15548997")
	return &NotificationHandler{
		MatchNotifier: MatchHandler,
		ErrorHandler:  ErrorHandler,
	}
}

func main() {
	Notifier := CreateNotificationHandler()

	qt := MatchFinder.TwoUp
	qR, err := MatchFinder.Find(qt)
	if err != nil {
		notificationErr := Notifier.LogError(err)
		if notificationErr != nil {
			errStr := fmt.Sprintf("Error logging error: %v\n", err)
			log.Println(errStr)
			return
		}
	}
	err = Notifier.NotifyUser(context.Background(), &qR)
	if err != nil {
		fmt.Println("Failed Webhook:", err)
		notificationErr := Notifier.LogError(err)
		if notificationErr != nil {
			errStr := fmt.Sprintf("Error logging error: %v\n", err)
			log.Println(errStr)
			return
		}
	}
}
