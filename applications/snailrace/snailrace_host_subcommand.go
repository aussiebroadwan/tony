package snailrace_app

import (
	"strings"

	"github.com/aussiebroadwan/tony/framework"
	"github.com/aussiebroadwan/tony/pkg/snailrace"
	"github.com/bwmarrin/discordgo"
)

type SnailraceHostSubCommand struct {
	framework.ApplicationSubCommand
	framework.ApplicationEvent
}

func (c SnailraceHostSubCommand) GetType() framework.AppType {
	return framework.AppTypeSubCommand | framework.AppTypeEvent
}

func (c SnailraceHostSubCommand) OnCommand(ctx framework.CommandContext) {

	err := snailrace.HostRace(stateRenderer(ctx))
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

func handleJoinRequest(ctx framework.EventContext, raceId string) {
	user := ctx.GetUser()

	snails, err := snailrace.GetSnails(user.ID)
	if err != nil {
		ctx.Logger().WithError(err).Error("Failed to get snails")
		return
	}

	// Convert user's snails to modal options
	menuOptions := make([]discordgo.SelectMenuOption, len(snails))
	for i, snail := range snails {
		menuOptions[i] = discordgo.SelectMenuOption{
			Label:   snail.Name, // TODO: Add usages left
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

	// You can react to button presses with no data and it doesn't error or send a message
	ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: nil,
	})
}
