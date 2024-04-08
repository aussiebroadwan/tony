package remind

import (
	"fmt"
	"time"

	"github.com/aussiebroadwan/tony/framework"
	"github.com/bwmarrin/discordgo"
)

// This is the subcommand for adding a reminder to the bot. The user can specify
// a time and message for the reminder and the bot will remind the user at the
// specified time.
//
//	/remind add <time> <message>
//
// An error will be returned if the time is not in the correct format of
// "2022-01-01 15:04:05".
type RemindAddSubCommand struct {
	framework.ApplicationSubCommand
}

func (c RemindAddSubCommand) GetType() framework.AppType {
	return framework.AppTypeSubCommand
}

func (c RemindAddSubCommand) OnCommand(ctx framework.CommandContext) {
	interaction := ctx.Interaction()
	db := ctx.Database()
	commandOptions := interaction.ApplicationCommandData().Options[0].Options

	// Get the time and message from the interaction
	triggerTimeStr, err := framework.GetOption(commandOptions, "time")
	if err != nil {
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "**Error:** Time is required",
			},
		})
		return
	}

	message, err := framework.GetOption(commandOptions, "message")
	if err != nil {
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "**Error:** Message is required",
			},
		})
		return
	}

	// Check if the time is valid
	triggerTime, err := time.ParseInLocation(time.DateTime, triggerTimeStr.StringValue(), time.Local)
	if err != nil {
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "**Error:** Invalid time format eg. 2022-01-01 15:04:05",
			},
		})
		return
	}

	// Get the user who created the reminder
	user := interaction.User
	if user == nil {
		user = interaction.Member.User
	}

	// Add the reminder
	id, err := AddReminder(
		db,
		user.Mention(),
		triggerTime,
		ctx.Session(),
		interaction.ChannelID,
		message.StringValue(),
	)

	if err != nil {
		ctx.Logger().WithError(err).Error("Failed to add reminder")
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "**Error:** Failed to add reminder",
			},
		})
		return
	}

	// Respond with the reminder ID
	ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: fmt.Sprintf("Reminder added `[%d]`", id),
		},
	})
}
