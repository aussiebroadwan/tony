package snailrace_app

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aussiebroadwan/tony/applications/snailrace/render"

	"github.com/aussiebroadwan/tony/framework"
	"github.com/aussiebroadwan/tony/pkg/snailrace"
	"github.com/aussiebroadwan/tony/pkg/tradingcards"
	"github.com/bwmarrin/discordgo"

	log "github.com/sirupsen/logrus"
)

type SnailraceHostSubCommand struct {
	framework.ApplicationSubCommand
	framework.ApplicationEvent
}

func (c SnailraceHostSubCommand) GetType() framework.AppType {
	return framework.AppTypeSubCommand | framework.AppTypeEvent
}

func (c SnailraceHostSubCommand) OnCommand(ctx framework.CommandContext) {

	err := snailrace.HostRace(render.StateRenderer(ctx))
	if err != nil {
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "Failed to host snailrace",
			},
		})
		return
	}

	// Respond with the reminder ID
	ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: "Setting up a snailrace...",
		},
	})
}

func (c SnailraceHostSubCommand) OnEvent(ctx framework.EventContext, eventType discordgo.InteractionType) {
	eventKey := ctx.EventValue()

	if eventType != discordgo.InteractionMessageComponent {
		ctx.Logger().Error("Invalid event type")
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "**Error**: Invalid event type",
			},
		})
	}

	values := strings.Split(eventKey, ":")

	switch values[0] {
	case "join_request":
		handleJoinRequest(ctx, values[1])
	case "join_select":
		handleJoin(ctx, values[1])
	default:
		ctx.Logger().Error("Invalid event key: " + values[0])
	}
}

func OnNewSnail(ctx framework.EventContext) func(snailrace.Snail) {
	return func(snail snailrace.Snail) {
		ctx.Logger().WithFields(log.Fields{
			"src":   "snailrace",
			"snail": snail.Id,

			"speed":        snail.Speed,
			"acceleration": snail.Acceleration,
			"stamina":      snail.Stamina,
			"weight":       snail.Weight,
			"luck":         snail.Luck,
		}).Info("Generated snail: ", snail.Name)

		err := tradingcards.RegisterCard(ctx.Database(), tradingcards.Card{
			Name: snail.Id,

			Title:       snail.Name,
			Description: fmt.Sprintf("A raceable %s snail.", snailrace.SnailType(snail.Type)),
			Application: "snailrace",

			Rarity:      tradingcards.CardRarityCommon,
			Usable:      true,
			Tradable:    true,
			Unbreakable: false,
			MaxUsage:    CommonSnailUsage,
			SVG:         "",
		})
		if err != nil {
			ctx.Logger().WithError(err).Error("Failed to register snail card")
			return
		}

		ctx.Logger().Info("Registered snail card: ", snail.Id)
		err = tradingcards.AssignCard(ctx.Database(), ctx.GetUser().ID, snail.Id)
		if err != nil {
			ctx.Logger().WithError(err).Error("Failed to assign snail card")
			return
		}
	}
}

func handleJoinRequest(ctx framework.EventContext, raceId string) {
	user := ctx.GetUser()

	snails, err := snailrace.GetSnails(user.ID, OnNewSnail(ctx))
	if err != nil {
		ctx.Logger().WithError(err).Error("Failed to get snails")
		return
	}

	// Convert user's snails to modal options
	menuOptions := make([]discordgo.SelectMenuOption, len(snails))
	for i, snail := range snails {
		// Get the trading card usages
		card, _ := tradingcards.GetUserCard(ctx.Database(), user.ID, snail.Id)
		menuOptions[i] = discordgo.SelectMenuOption{
			Label:   fmt.Sprintf("%s (%d/%d)", snail.Name, card.CurrentUsage, card.MaxUsage),
			Value:   snail.Id,
			Default: false,
		}
	}

	// Handle join request
	err = ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:    discordgo.MessageFlagsEphemeral,
			CustomID: "snailrace.host:join:" + raceId,
			Title:    "Join Snailrace",
			Content:  "Select a snail from your deck to join in the race",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "snailrace.host:join_select:" + raceId,
							Placeholder: "Select a snail",
							Options:     menuOptions,
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

func handleJoin(ctx framework.EventContext, raceId string) {
	data := ctx.Interaction().MessageComponentData()
	snail := data.Values[0]

	err := snailrace.JoinRace(ctx.GetUser().ID, raceId, snail)
	if err != nil {
		ctx.Logger().WithError(err).Errorf("User %s has failed to join race %s using snail %s", ctx.GetUser().Username, raceId, snail)
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "**Error**: " + err.Error(),
			},
		})
		return
	}

	// Decrement the snail's usage
	err = tradingcards.UseCard(ctx.Database(), ctx.GetUser().ID, snail)
	if errors.Is(err, tradingcards.ErrCardBroken) {
		// You can react to button presses with no data and it doesn't error or send a message
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "This will be the last time you can race this snail. It will be removed from your deck on race completion.",
			},
		})
	}

	// You can react to button presses with no data and it doesn't error or send a message
	ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: nil,
	})
}
