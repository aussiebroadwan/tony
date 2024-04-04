package applications

import (
	"fmt"

	"github.com/aussiebroadwan/tony/framework"
	"github.com/bwmarrin/discordgo"
)

// PingCommand defines a command for responding to "ping" interactions
// with a simple "Pong!" message. This command demonstrates a basic
// interaction within Discord using the discordgo package.
type PingCommand struct {
	framework.Command
}

// Register is responsible for registering the "ping" command with
// Discord's API. It defines the command name and description that
// appear in the Discord user interface.
func (pc *PingCommand) Register(s *discordgo.Session) *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "ping",
		Description: "Replies with Pong!",
	}
}

func (pc *PingCommand) GetType() framework.CommandType {
	return framework.CommandTypeApp
}

// Execute handles the execution logic for the "ping" command. When a user
// invokes this command, Discord triggers this method, allowing the bot to
// respond appropriately.
func (pc *PingCommand) Execute(ctx *framework.Context) {
	interaction := ctx.Interaction()

	user := interaction.Member.User
	if user == nil {
		user = interaction.User
	}

	ctx.Session().InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Pong %s!", user.Mention()),
		},
	})
}

// PingButtonCommand defines a subcommand for responding to "ping.button"
// interactions with a button that the user can interact with. This command
// demonstrates how to create a button that users can click to interact with
// the bot.
type PingButtonCommand struct {
	framework.SubCommand
}

func (pc *PingButtonCommand) GetType() framework.CommandType {
	return framework.CommandTypeAppAndEvent
}

// Register is responsible for registering the "ping.button" command with
// Discord's API. It defines the command name and description that appear
// in the Discord user interface.
func (pc *PingButtonCommand) Register(s *discordgo.Session) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "button",
		Description: "Makes a button for you to ping the bot!",
	}
}

// Execute handles the execution logic for the "ping.button" command. When a
// user invokes this command, Discord triggers this method, allowing the bot
// to post a button that the user can interact with.
func (pc *PingButtonCommand) Execute(ctx *framework.Context) {
	err := ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label: "Ping",
							Style: discordgo.SuccessButton,
							Emoji: &discordgo.ComponentEmoji{
								Name: "🏓",
							},
							CustomID: "ping.button:ping",
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

// OnEvent handles the event logic for the "ping.button" command. When a user
// interacts with the button, Discord triggers this method, allowing the bot
// to respond appropriately.
func (pc *PingButtonCommand) OnEvent(ctx *framework.Context, eventType discordgo.InteractionType) {
	value := ctx.EventValue()
	interaction := ctx.Interaction()

	// Get the user from the interaction
	user := interaction.Member.User
	if user == nil {
		user = interaction.User
	}

	// Reject non message component interactions
	if eventType != discordgo.InteractionMessageComponent {
		ctx.Session().InteractionRespond(interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("I only respond to button presses %s :(", user.Mention()),
			},
		})
		return
	}

	// Reject non-ping button presses
	if value != "ping" {
		ctx.Session().InteractionRespond(interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("I only respond to the ping button %s :(", user.Mention()),
			},
		})
		return
	}

	// Respond to the interaction Successfully
	ctx.Session().InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Pong %s!", user.Mention()),
		},
	})
}
