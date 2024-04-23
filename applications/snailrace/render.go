package snailrace_app

import (
	"github.com/aussiebroadwan/tony/framework"
	"github.com/aussiebroadwan/tony/pkg/snailrace"
	"github.com/aussiebroadwan/tony/pkg/wallet"
	"github.com/bwmarrin/discordgo"
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

	return createGameStateRenderFunc(ctx, session, creditUser), onAchievement(ctx), interaction.ChannelID, msg.ID
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
		ctx.Logger().WithField("state", raceState.State).Info("Rendering game state")
		session.ChannelMessageEdit(channelId, messageId, preparingGameMessage)
	}
}
