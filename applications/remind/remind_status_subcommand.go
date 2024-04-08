package remind

import (
	"fmt"

	"github.com/aussiebroadwan/tony/framework"
	"github.com/bwmarrin/discordgo"
)

// This is the subcommand for checking the status of a reminder. The user can
// specify the ID of the reminder to check and the bot will respond with the
// time left until the reminder is triggered. It will only respond if the user
// owns the reminder.
//
//	/remind status <id>
//
// If the reminder is not found, an error will be returned.
type RemindStatusSubCommand struct {
	framework.ApplicationSubCommand
}

func (c RemindStatusSubCommand) GetType() framework.AppType {
	return framework.AppTypeSubCommand
}

func (c RemindStatusSubCommand) OnCommand(ctx framework.CommandContext) {
	interaction := ctx.Interaction()
	commandOptions := interaction.ApplicationCommandData().Options[0].Options

	// Get the reminder ID from the interaction
	id, err := framework.GetOption(commandOptions, "id")
	if err != nil {
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "**Error:** id is required",
			},
		})
		return
	}

	// Get the reminder status
	timeLeft, err := Status(id.IntValue())
	if err != nil {
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: fmt.Sprintf("**Error:** reminder `[%d]` not found", id.IntValue()),
			},
		})
		return
	}

	// Respond with the reminder status
	ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: fmt.Sprintf("Time left for `[%d]`: `%s`", id.IntValue(), timeLeft.String()),
		},
	})
}
