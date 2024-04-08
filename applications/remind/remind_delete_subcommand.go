package remind

import (
	"fmt"

	"github.com/aussiebroadwan/tony/framework"
	"github.com/bwmarrin/discordgo"
)

// This is the subcommand for deleting a reminder from the bot. The user can
// specify the ID of the reminder to delete and the bot will remove the reminder
// from the list of reminders.
//
//	/remind del <id>
//
// If the reminder is not found, an error will be returned. If the user is not
// the creator of the reminder, an error will be returned.
type RemindDeleteSubCommand struct {
	framework.ApplicationSubCommand
}

func (c RemindDeleteSubCommand) GetType() framework.AppType {
	return framework.AppTypeSubCommand
}

func (c RemindDeleteSubCommand) OnCommand(ctx framework.CommandContext) {
	interaction := ctx.Interaction()
	db := ctx.Database()
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

	user := interaction.User
	if user == nil {
		user = interaction.Member.User
	}

	// Delete the reminder
	err = DeleteReminder(db, id.IntValue(), user.Mention())
	if err != nil {
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: fmt.Sprintf("**Error:** reminder `[%d]` not found", id.IntValue()),
			},
		})
	}

	// Respond that the reminder was deleted
	ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: fmt.Sprintf("Reminder `[%d]` deleted", id.IntValue()),
		},
	})
}
