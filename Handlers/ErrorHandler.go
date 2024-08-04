package Handlers

import "Adam/discord-twoup/Discord"

type ErrorHandler interface {
	LogError(error) error
}

type DiscordErrorHandler struct {
	Url   string
	Color string
}

func (d *DiscordErrorHandler) LogError(err error) error {
	title := "Error"
	ErrMsg := err.Error()
	inline := false
	fields := []webhook_manager.Field{
		{
			Name:   "Message",
			Value:  &ErrMsg,
			Inline: &inline,
		},
	}
	embed := []webhook_manager.Embed{{
		Title:  &title,
		Color:  &d.Color,
		Fields: &fields,
	}}

	Msg := webhook_manager.Message{
		Content:         nil,
		Embeds:          &embed,
		AllowedMentions: nil,
	}

	discErr := webhook_manager.SendMessage(d.Url, Msg)
	if discErr != nil {
		return discErr
	}

	return nil
}

func NewDiscordErrorHandler(url string, color string) *DiscordErrorHandler {
	return &DiscordErrorHandler{
		Url:   url,
		Color: color,
	}
}
