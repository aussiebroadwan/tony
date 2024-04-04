package commands

import (
	"fmt"
	"time"

	"github.com/aussiebroadwan/tony/database"
	"github.com/aussiebroadwan/tony/framework"
	"github.com/aussiebroadwan/tony/pkg/reminders"
	"github.com/bwmarrin/discordgo"
)

type RemindCommand struct {
	framework.Command
}

// Register is responsible for registering the "remind" command with
// Discord's API. It defines the command name and description that
// appear in the Discord user interface.
func (c *RemindCommand) Register(s *discordgo.Session) *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "remind",
		Description: "Allows users to set reminders",
	}
}

func (c *RemindCommand) Execute(ctx *framework.Context) {
	// NOP
	//
	// This command is a parent command and does not have any  execution logic.
	// It is used to group subcommands.
}

func (c *RemindCommand) OnEvent(ctx *framework.Context, eventType discordgo.InteractionType) {
	/* NOP */
}

// This is the subcommand for adding a reminder to the bot. The user can specify
// a time and message for the reminder and the bot will remind the user at the
// specified time.
//
//	/remind add <time> <message>
//
// An error will be returned if the time is not in the correct format of
// "2022-01-01 15:04:05".
type RemindAddSubCommand struct {
	framework.SubCommand
}

func (c *RemindAddSubCommand) Register(s *discordgo.Session) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "add",
		Description: "Add a reminder",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message",
				Description: "The message to remind you about",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "time",
				Description: "The time to remind you",
				Required:    true,
			},
		},
	}
}

func (c *RemindAddSubCommand) Execute(ctx *framework.Context) {
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
	id, err := database.AddReminder(
		db,
		user.Mention(),
		triggerTime,
		ctx.Session(),
		interaction.ChannelID,
		message.StringValue(),
	)

	if err != nil {
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

func (com *RemindAddSubCommand) OnEvent(ctx *framework.Context, eventType discordgo.InteractionType) {
	/* NOP */
}

// This is the subcommand for deleting a reminder from the bot. The user can
// specify the ID of the reminder to delete and the bot will remove the reminder
// from the list of reminders.
//
//	/remind del <id>
//
// If the reminder is not found, an error will be returned. If the user is not
// the creator of the reminder, an error will be returned.
type RemindDeleteSubCommand struct {
	framework.SubCommand
}

func (c *RemindDeleteSubCommand) Register(s *discordgo.Session) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "del",
		Description: "Delete a reminder",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "id",
				Description: "The ID of the reminder to delete",
				Required:    true,
			},
		},
	}
}

func (c *RemindDeleteSubCommand) Execute(ctx *framework.Context) {
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
	err = database.DeleteReminder(db, id.IntValue(), user.Mention())
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

func (c *RemindDeleteSubCommand) OnEvent(ctx *framework.Context, eventType discordgo.InteractionType) {
	/* NOP */
}

// This is the subcommand for listing all reminders for the user. The user can
// view all reminders that they have created and are able to delete or check
// the status of.
//
//	/remind list
//
// If the user has no reminders, the bot will respond with "No reminders found".
type RemindListSubCommand struct {
	framework.SubCommand
}

func (c *RemindListSubCommand) Register(s *discordgo.Session) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "list",
		Description: "List all reminders",
	}
}

func (c *RemindListSubCommand) Execute(ctx *framework.Context) {
	interaction := ctx.Interaction()

	// Get all reminders
	reminderList := reminders.List()

	// Get the user who created the reminder
	user := interaction.User
	if user == nil {
		user = interaction.Member.User
	}

	// Fetch current user's reminders
	var userReminders = make([]reminders.Reminder, 0)
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

func (c *RemindListSubCommand) OnEvent(ctx *framework.Context, eventType discordgo.InteractionType) {
	/* NOP */
}

// This is the subcommand for checking the status of a reminder. The user can
// specify the ID of the reminder to check and the bot will respond with the
// time left until the reminder is triggered. It will only respond if the user
// owns the reminder.
//
//	/remind status <id>
//
// If the reminder is not found, an error will be returned.
type RemindStatusSubCommand struct {
	framework.SubCommand
}

func (c *RemindStatusSubCommand) Register(s *discordgo.Session) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "status",
		Description: "Get the status of a reminder",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "id",
				Description: "The ID of the reminder to check",
				Required:    true,
			},
		},
	}
}

func (c *RemindStatusSubCommand) Execute(ctx *framework.Context) {
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
	timeLeft, err := reminders.Status(id.IntValue())
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

func (c *RemindStatusSubCommand) OnEvent(ctx *framework.Context, eventType discordgo.InteractionType) {
	/* NOP */
}
