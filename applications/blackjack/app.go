package blackjack

import (
	"fmt"
	"strings"

	"github.com/aussiebroadwan/tony/framework"
	"github.com/bwmarrin/discordgo"
)

func RegisterBlackjackApp(bot *framework.Bot) framework.Route {
	return framework.NewRoute(bot, "blackjack",
		&Blackjack{},
	)
}

type Blackjack struct {
	framework.ApplicationCommand
	framework.ApplicationEvent
}

func (b Blackjack) GetType() framework.AppType {
	return framework.AppTypeCommand | framework.AppTypeEvent
}

func (b Blackjack) GetDefinition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "blackjack",
		Description: "Let's play some blackjack!",
	}
}

func (b Blackjack) OnCommand(ctx framework.CommandContext) {
	// 1. Check if there is currently a game in progress
	// 2. If there is a game in progress, send a message to the user

	// 3. If there is no game in progress, start a new game
	err := ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Let's play some blackjack!",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label: "Join",
							Style: discordgo.SuccessButton,
							Emoji: &discordgo.ComponentEmoji{
								Name: "🃏",
							},
							CustomID: "blackjack:host",
						},
					},
				},
			},
		},
	})
	if err != nil {
		ctx.Logger().WithError(err).Error("Failed to respond to interaction")
	}
}

const (
	HostEvent  string = "host-" + string(discordgo.InteractionMessageComponent)
	JoinEvent  string = "join-" + string(discordgo.InteractionModalSubmit)
	HitEvent   string = "hit-" + string(discordgo.InteractionMessageComponent)
	StandEvent string = "stand-" + string(discordgo.InteractionMessageComponent)
)

func (b Blackjack) OnEvent(ctx framework.EventContext, eventType discordgo.InteractionType) {
	eventKey, eventValue, _ := strings.Cut(ctx.EventValue(), ":")
	eventKey += "-" + string(eventType)

	switch eventKey {
	case HostEvent:
		OnHost(ctx)
	case JoinEvent: // This is a Modal Submit event
		OnJoin(ctx, eventValue)
		return
	case HitEvent:
		OnHit(ctx)
	case StandEvent:
		OnStand(ctx)
	default:
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: fmt.Sprintf("Unknown event %s", eventKey),
			},
		})
	}

	// You can react to button presses with no data and it doesn't error or send a message
	ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: nil,
	})
}

func OnHost(ctx framework.EventContext) {
}

func OnJoin(ctx framework.EventContext, eventValue string) {
}

func OnHit(ctx framework.EventContext) {
}

func OnStand(ctx framework.EventContext) {
}
