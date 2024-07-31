package main

import (
	webhook_manager "Adam/discord-twoup/Discord"
	"Adam/discord-twoup/MatchFinder"
	"fmt"
	"log"
)

const url = "https://canary.discord.com/api/webhooks/1267996258203865201/ZenhYVAIrjy3YLLkqtAjSIEGJ3W6UW2Qn627G41IrcxiaAZrSTGjegYR9zDbwfUkmD4v"

func main() {

	qt := MatchFinder.TwoUp
	qR, err := MatchFinder.Find(qt)
	if err != nil {
		err := webhook_manager.SendMessage(url, createPlainMessage(*err))
		if err != nil {
			errStr := fmt.Sprintf("Error logging error: %v\n", err)
			log.Println(errStr)
			return
		}
	}
	var Embed []webhook_manager.Embed
	for _, v := range qR {
		Embed = append(Embed, createOddsEmbed(v))
	}
	baseMessage := createSilentMessage(Embed)
	whError := webhook_manager.SendMessage(url, baseMessage)
	if err != nil {
		fmt.Println("Failed Webhook:", err)
		err := webhook_manager.SendMessage(url, logError(whError))
		if err != nil {
			errStr := fmt.Sprintf("Error logging error: %v\n", err)
			log.Println(errStr)
			return
		}
	}
}

func createOddsEmbed(match MatchFinder.Match) webhook_manager.Embed {
	var inline = true
	var colour = "1207889"
	Fields := []webhook_manager.Field{
		{
			Name:   `Event Name`,
			Value:  &match.EventName,
			Inline: &inline,
		},
		{
			Name:   `Team Selection`,
			Value:  &match.SelectionName,
			Inline: &inline,
		},
		{
			Name:   `Back Odds`,
			Value:  &match.Back.Odds,
			Inline: &inline,
		},
		{
			Name:   `Lay Odds`,
			Value:  &match.Lay.Odds,
			Inline: &inline,
		},
		{
			Name:   `Rating`,
			Value:  &match.Rating,
			Inline: &inline,
		},
	}

	embed := webhook_manager.Embed{
		Title:  &match.EventName,
		Color:  &colour,
		Fields: &Fields,
	}
	return embed
}

func createSilentMessage(embeds []webhook_manager.Embed) webhook_manager.Message {
	var groups = []string{"261512678269255681"}
	var content = "<@261512678269255681>"
	var allowedMentions = webhook_manager.AllowedMentions{Users: &groups}
	message := webhook_manager.Message{
		Content:         &content,
		Embeds:          &embeds,
		AllowedMentions: &allowedMentions,
	}

	return message
}

func createPlainMessage(content string) webhook_manager.Message {
	message := webhook_manager.Message{
		Content: &content,
	}
	return message
}

func logError(err error) webhook_manager.Message {
	return createPlainMessage(err.Error())
}
