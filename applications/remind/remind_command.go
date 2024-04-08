package remind

import (
	"github.com/aussiebroadwan/tony/framework"
	"github.com/bwmarrin/discordgo"
)

func RegisterRemindApp(bot *framework.Bot) framework.Route {
	return framework.NewRoute(bot, "remind",
		// remind
		&RemindCommand{}, // [NOP]

		// remind <subcommand>
		framework.NewRoute(bot, "add", &RemindAddSubCommand{}),
		framework.NewRoute(bot, "del", &RemindDeleteSubCommand{}),
		framework.NewRoute(bot, "list", &RemindListSubCommand{}),
		framework.NewRoute(bot, "status", &RemindStatusSubCommand{}),
	)
}

type RemindCommand struct {
	framework.ApplicationCommand
}

func (c RemindCommand) GetType() framework.AppType {
	return framework.AppTypeCommand | framework.AppTypeMountable
}

func (c RemindCommand) OnMount(ctx framework.MountContext) {

	// Setup reminders
	go Run()
	SetupRemindersDB(ctx.Database(), ctx.Session())
}

// Register is responsible for registering the "remind" command with
// Discord's API. It defines the command name and description that
// appear in the Discord user interface.
func (c RemindCommand) GetDefinition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "remind",
		Description: "Allows users to set reminders",
		Options: []*discordgo.ApplicationCommandOption{
			{
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
			},
			{
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
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "list",
				Description: "List all reminders",
			},
			{
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
			},
		},
	}
}

func (c RemindCommand) OnCommand(ctx framework.CommandContext) {
	// This is a NOP command and should not be executed directly
}
