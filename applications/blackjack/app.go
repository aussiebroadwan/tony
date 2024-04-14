package blackjackApp

import (
	"slices"

	"github.com/aussiebroadwan/tony/framework"
	"github.com/aussiebroadwan/tony/pkg/blackjack"
	"github.com/bwmarrin/discordgo"
)

var (
	HostEvent  string = "host-" + discordgo.InteractionMessageComponent.String()
	JoinEvent  string = "join-" + discordgo.InteractionModalSubmit.String()
	HitEvent   string = "hit-" + discordgo.InteractionMessageComponent.String()
	StandEvent string = "stand-" + discordgo.InteractionMessageComponent.String()
)

func RegisterBlackjackApp(bot *framework.Bot) framework.Route {
	return framework.NewRoute(bot, "blackjack", &Blackjack{})
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
	if blackjack.Running() {
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "A game is already in progress",
			},
		})
		return
	}

	err := blackjack.Host(stateRenderer(ctx))
	if err != nil {
		ctx.Logger().WithError(err).Error("Failed to start a game")
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "Failed to start a game",
			},
		})
		return
	}

	// You can react to button presses with no data and it doesn't error or send a message
	err = ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: "Booting up Blackjack..",
		},
	})
	if err != nil {
		ctx.Logger().WithError(err).Error("Failed to respond to interaction")
	}

	ctx.Logger().Info("Blackjack game started")
}

func (b Blackjack) OnEvent(ctx framework.EventContext, eventType discordgo.InteractionType) {
	eventKey := ctx.EventValue() + "-" + eventType.String()

	// Check if the event key is valid
	if !slices.Contains([]string{HostEvent, JoinEvent, HitEvent, StandEvent}, eventKey) {
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "A game is already in progress",
			},
		})
	}

	// Switch on the event key
	switch eventKey {
	case HostEvent: // Returns a message with a joining button that builds a modal
		OnHost(ctx)
	case JoinEvent: // This is a Modal Submit event
		OnJoin(ctx)
	case HitEvent:
		OnHit(ctx)
	case StandEvent:
		OnStand(ctx)
	}
}

func OnHost(ctx framework.EventContext) {
	// You can react to button presses with no data and it doesn't error or send a message
	err := ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "blackjack:join",
			Title:    "Join Blackjack",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "bet",
							Label:       "Bet Amount",
							Style:       discordgo.TextInputShort,
							Placeholder: "eg. 15",
							Required:    true,
							MaxLength:   3,
							MinLength:   2,
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

func OnJoin(ctx framework.EventContext) {

	data := ctx.Interaction().ModalSubmitData()
	bet := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

	ctx.Logger().Infof("User %s has joined with a bet of %s", ctx.GetUser().Username, bet)

	// You can react to button presses with no data and it doesn't error or send a message
	err := ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: nil,
	})
	if err != nil {
		ctx.Logger().WithError(err).Error("Failed to respond to interaction")
	}
}

func OnHit(ctx framework.EventContext) {
	// You can react to button presses with no data and it doesn't error or send a message
	ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: nil,
	})
}

func OnStand(ctx framework.EventContext) {
	// You can react to button presses with no data and it doesn't error or send a message
	ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: nil,
	})
}
