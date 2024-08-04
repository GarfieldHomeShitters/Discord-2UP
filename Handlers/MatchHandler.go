package Handlers

import (
	"Adam/discord-twoup/Discord"
	"Adam/discord-twoup/MatchFinder"
	"bytes"
	"context"
	"fmt"
	"time"
)

type MatchNotificationError struct {
	err string
}

func (e MatchNotificationError) Error() string {
	return fmt.Sprintf("[MatchNotificationError] %s", e.err)
}

type MatchNotifier interface {
	NotifyUser(context context.Context, Matches *[]MatchFinder.Match) error
}

type DiscordMatchNotifier struct {
	Url   string
	Ping  DiscordPing
	Color string
}

type DiscordPing struct {
	ParseStrings []string
	UserIDs      []string
	RoleIDs      []string
}

func (n *DiscordMatchNotifier) NotifyUser(ctx context.Context, Matches *[]MatchFinder.Match) error {
	if len(*Matches) == 0 {
		return fmt.Errorf("no_matches")
	}

	embeds := n.CreateEmbeds(Matches)
	content := n.CreateContent()
	mentions, err := n.CreateAllowedMentions()
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	chn := make(chan error)

	go func() {
		err := webhook_manager.SendMessage(n.Url,
			webhook_manager.Message{
				Content:         &content,
				Embeds:          &embeds,
				AllowedMentions: mentions,
			},
		)
		chn <- err
	}()

	for {
		select {
		case <-chn:
			return err
		case <-ctx.Done():
			return MatchNotificationError{err: ctx.Err().Error()}
		}
	}
}

func (n *DiscordMatchNotifier) CreateEmbeds(Matches *[]MatchFinder.Match) []webhook_manager.Embed {
	inline, colour := true, n.Color
	var Embeds []webhook_manager.Embed

	for _, match := range *Matches {
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

		Embeds = append(Embeds,
			webhook_manager.Embed{
				Title:  &match.EventName,
				Color:  &colour,
				Fields: &Fields,
			})
	}

	return Embeds
}

func (n *DiscordMatchNotifier) CreateContent() string {
	strBuf := &bytes.Buffer{}
	for i, v := range n.Ping.ParseStrings {
		if i > 0 {
			strBuf.WriteByte(' ')
		}
		strBuf.WriteString("<@")
		strBuf.WriteString(v)
		strBuf.WriteString(">")
	}
	for i, v := range n.Ping.UserIDs {
		if i > 0 {
			strBuf.WriteByte(' ')
		}
		strBuf.WriteString("<@")
		strBuf.WriteString(v)
		strBuf.WriteString(">")
	}
	for i, v := range n.Ping.RoleIDs {
		if i > 0 {
			strBuf.WriteByte(' ')
		}
		strBuf.WriteString("<@")
		strBuf.WriteString(v)
		strBuf.WriteString(">")
	}
	return strBuf.String()
}

func (n *DiscordMatchNotifier) CreateAllowedMentions() (*webhook_manager.AllowedMentions, error) {
	return &webhook_manager.AllowedMentions{Parse: &n.Ping.ParseStrings, Users: &n.Ping.UserIDs, Roles: &n.Ping.RoleIDs}, nil
}

func NewPingType(parse []string, users []string, roles []string) *DiscordPing {
	return &DiscordPing{
		ParseStrings: parse,
		UserIDs:      users,
		RoleIDs:      roles,
	}
}

func NewDiscordMatchNotifier(url string, ping DiscordPing, colour string) *DiscordMatchNotifier {
	return &DiscordMatchNotifier{
		Url:   url,
		Ping:  ping,
		Color: colour,
	}
}
