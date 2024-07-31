package main

import (
	webhook_manager "Adam/discord-twoup/Discord"
	"Adam/discord-twoup/MatchFinder"
)

func main() {

	qt := MatchFinder.TwoUp
	qR := MatchFinder.Find(qt)

	baseMessage := createSilentMessage()
	webhook_manager.SendMessage("https://canary.discord.com/api/webhooks/1267996258203865201/ZenhYVAIrjy3YLLkqtAjSIEGJ3W6UW2Qn627G41IrcxiaAZrSTGjegYR9zDbwfUkmD4v", baseMessage)

}

func createEmbed() webhook_manager.Embed {
	embed := webhook_manager.Embed{}
	return embed
}

func createSilentMessage() webhook_manager.Message {
	var groups = []string{"@everyone"}
	var content = "[](@everyone)"
	var allowedMentions = webhook_manager.AllowedMentions{Parse: &groups}
	message := webhook_manager.Message{
		Content:         &content,
		Embeds:          nil,
		AllowedMentions: &allowedMentions,
	}

	return message
}
