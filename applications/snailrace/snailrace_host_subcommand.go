package snailrace_app

import (
	"fmt"

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

	if eventKey == "join_request" && eventType == discordgo.InteractionMessageComponent {
		// Handle join request
		handleJoinRequest(ctx)
		return
	}

	if eventKey == "join" && eventType == discordgo.InteractionModalSubmit {
		// Handle join
		handleJoin(ctx)
		return
	}
}

func handleJoinRequest(ctx framework.EventContext) {
	user := ctx.GetUser()

	snails, err := snailrace.GetSnails(user.ID)
	if err != nil {
		ctx.Logger().WithError(err).Error("Failed to get snails")
		return
	}

	// Convert user's snails to modal options
	modalOptions := make([]discordgo.SelectMenuOption, len(snails))
	for i, snail := range snails {
		modalOptions[i] = discordgo.SelectMenuOption{
			Label:   snail.Name,
			Value:   snail.Id,
			Default: false,
		}
	}

	// Handle join request
	minValues := 1
	err = ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:    discordgo.MessageFlagsEphemeral,
			CustomID: "snailrace.host:join",
			Title:    "Join Snailrace",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "snailrace.host:join_select",
							Placeholder: "Select a snail",
							Options:     modalOptions,
							MinValues:   &minValues,
							MaxValues:   1,
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

func handleJoin(ctx framework.EventContext) {
	data := ctx.Interaction().ModalSubmitData()
	menu := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.SelectMenu)
	fmt.Printf("%+v\n", menu)

	// ctx.Logger().Infof("User %s has joined using snail %s", ctx.GetUser().Username, snail)
}
