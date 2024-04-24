package snailrace_app

import (
	"fmt"
	"time"

	"github.com/aussiebroadwan/tony/framework"
	"github.com/aussiebroadwan/tony/pkg/snailrace"
	"github.com/aussiebroadwan/tony/pkg/wallet"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// Constants for embed colors and game messages.
const (
	embedColor           = 0x034FA48
	preparingGameMessage = "Snailrace game in preparing"
)

// stateRenderer sets up the messaging and functionality for rendering game states in blackjack.
func stateRenderer(ctx framework.CommandContext) (snailrace.StateChangeCallback, snailrace.AchievementCallback, string, string) {
	session := ctx.Session()
	interaction := ctx.Interaction()
	database := ctx.Database()

	msg, err := session.ChannelMessageSend(interaction.ChannelID, preparingGameMessage)
	if err != nil {
		ctx.Logger().WithError(err).Error("Failed to create game message")
		return nil, nil, "", ""
	}

	creditUser := func(userId string, amount int64) {
		if err := wallet.Credit(database, userId, amount, "Snailrace returns", "snailrace"); err != nil {
			ctx.Logger().WithError(err).Error("Failed to credit user")
		}
	}

	return createGameStateRenderFunc(ctx, session, creditUser), onAchievement(ctx), msg.ID, interaction.ChannelID
}

// onAchievement creates a function to handle achievement unlocks. It will
// assign a card to the user if they unlock an achievement. If it fails to
// assign the card it will return false, notifying the game to try again when
// the condition is met.
func onAchievement(ctx framework.CommandContext) snailrace.AchievementCallback {
	return func(userId string, achievement string) bool {
		ctx.Logger().WithField("user", userId).Info("Achievement unlocked: " + achievement)
		return true
	}
}

// createGameStateRenderFunc creates a function to render the game state based on the current stage.
func createGameStateRenderFunc(ctx framework.CommandContext, session *discordgo.Session, creditUser func(string, int64)) snailrace.StateChangeCallback {
	return func(raceState snailrace.RaceState, messageId, channelId string) {
		ctx.Logger().WithFields(logrus.Fields{
			"state":   raceState.State,
			"channel": channelId,
			"message": messageId,
		}).Info("Rendering game state")

		var err error
		description := ""
		var components []discordgo.MessageComponent = nil

		description, components = joinMessage(raceState)

		// switch raceState.State {
		// case snailrace.StateJoining:
		// 	description, components = joinMessage(raceState)
		// case snailrace.StateBetting:
		// case snailrace.StateInProgress:
		// case snailrace.StateFinished:
		// default:
		// 	session.ChannelMessageEdit(channelId, messageId, preparingGameMessage)
		// 	return
		// }

		err = renderState(session, channelId, messageId, "Snailrace: "+raceState.Race.Id, description, components)

		if err != nil {
			ctx.Logger().WithField("state", raceState.State).WithError(err).Error("Failed to render game state")
		}
	}
}

// renderState updates the game message with new state information and interaction components.
func renderState(session *discordgo.Session, channelId, messageId, title, description string, components []discordgo.MessageComponent) error {
	edit := discordgo.NewMessageEdit(channelId, messageId)
	edit.Embeds = &[]*discordgo.MessageEmbed{{Title: title, Description: description, Color: embedColor}}
	if components != nil {
		edit.Components = &[]discordgo.MessageComponent{discordgo.ActionsRow{Components: components}}
	}
	_, err := session.ChannelMessageEditComplex(edit)
	return err
}

// joinMessage generates the join stage message and components.
func joinMessage(state snailrace.RaceState) (string, []discordgo.MessageComponent) {
	description := fmt.Sprintf(
		"A new race has been hosted!\n\nRace ID: `%s`\nStarting: %s\n\nClick the `Join` button to join with your own snail.\n**Entrants:**\n",
		state.Race.Id,
		state.Race.StartAt.Format(time.DateTime),
	)

	// Add the snails to the body as entrants `- <snail_name>(<@owner_id>)`
	for _, snail := range state.Snails {
		description += fmt.Sprintf("- %s\n", snail.Name)
	}

	return description, []discordgo.MessageComponent{
		discordgo.Button{
			Label:    "Join",
			Style:    discordgo.SuccessButton,
			CustomID: "snailrace.host:join_request",
		},
	}
}
