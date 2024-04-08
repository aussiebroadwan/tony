package remind

import (
	"fmt"

	"github.com/aussiebroadwan/tony/framework"
	"github.com/bwmarrin/discordgo"
)

// This is the subcommand for listing all reminders for the user. The user can
// view all reminders that they have created and are able to delete or check
// the status of.
//
//	/remind list
//
// If the user has no reminders, the bot will respond with "No reminders found".
type RemindListSubCommand struct {
	framework.ApplicationSubCommand
}

func (c RemindListSubCommand) GetDefinition() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "list",
		Description: "List all reminders",
	}
}

func (c RemindListSubCommand) GetType() framework.AppType {
	return framework.AppTypeSubCommand
}

func (c RemindListSubCommand) OnCommand(ctx framework.CommandContext) {
	interaction := ctx.Interaction()

	// Get all reminders
	reminderList := List()

	// Get the user who created the reminder
	user := interaction.User
	if user == nil {
		user = interaction.Member.User
	}

	// Fetch current user's reminders
	var userReminders = make([]Reminder, 0)
	for _, reminder := range reminderList {
		if reminder.CreatedBy == user.Mention() {
			userReminders = append(userReminders, reminder)
		}
	}

	// Respond with the reminder list
	if len(userReminders) == 0 {
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "No reminders found",
			},
		})
		return
	}

	var reminderListStr string = "Reminders:\n\n```\n"
	for _, reminder := range userReminders {
		reminderListStr += fmt.Sprintf("[%d]: %s\n", reminder.ID, reminder.TriggerTime.String())
	}
	reminderListStr += "```"

	ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: reminderListStr,
		},
	})
}
