package applications

import (
	"fmt"

	"github.com/aussiebroadwan/tony/framework"
	"github.com/bwmarrin/discordgo"
)

func RegisterPingApp(bot *framework.Bot) framework.Route {
	return framework.NewRoute(bot, "ping",
		// ping
		&PingCommand{}, // [NOP]
	)
}

// PingCommand defines a command for responding to "ping" interactions
// with a simple "Pong!" message. This command demonstrates a basic
// interaction within Discord using the discordgo package.
type PingCommand struct {
	framework.ApplicationCommand
}

func (pc PingCommand) GetType() framework.AppType {
	return framework.AppTypeCommand | framework.AppTypeEvent
}

// Register is responsible for registering the "ping" command with
// Discord's API. It defines the command name and description that
// appear in the Discord user interface.
func (pc PingCommand) GetDefinition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "ping",
		Description: "Replies with a button to Ping-Pong!",
	}
}

// OnCommand handles the execution logic for the "ping" command. When a user
// invokes this command, Discord triggers this method, allowing the bot to
// respond appropriately.
func (pc PingCommand) OnCommand(ctx framework.CommandContext) {
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
							CustomID: "ping:ping",
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

// OnEvent handles the event logic for the "ping" command. When a user
// interacts with the button, Discord triggers this method, allowing the bot
// to respond appropriately.
func (p PingCommand) OnEvent(ctx framework.EventContext, eventType discordgo.InteractionType) {
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
