package render

import (
	"fmt"
	"math"
	"strings"

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

	RaceLength = 100.0
	TrackWidth = 20.0
)

// stateRenderer sets up the messaging and functionality for rendering game states in blackjack.
func StateRenderer(ctx framework.CommandContext) (snailrace.StateChangeCallback, snailrace.AchievementCallback, string, string) {
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

		switch raceState.State {
		case snailrace.StateJoining:
			description, components = joinMessage(raceState)
		case snailrace.StateBetting:
			description, components = bettingMessage(raceState)
		case snailrace.StateInProgress:
			description, components = raceMessage(raceState)
		case snailrace.StateFinished:
			description, components = finishedMessage(raceState, creditUser)
		case snailrace.StateCancelled:
			description, components = cancelledMessage(raceState)
		default:
			session.ChannelMessageEdit(channelId, messageId, preparingGameMessage)
			return
		}

		err = renderState(session, channelId, messageId, "Snailrace: User Hosted", description, components)

		if err != nil {
			ctx.Logger().WithField("state", raceState.State).WithError(err).Error("Failed to render game state")
		}
	}
}

// renderState updates the game message with new state information and interaction components.
func renderState(session *discordgo.Session, channelId, messageId, title, description string, components []discordgo.MessageComponent) error {
	edit := discordgo.NewMessageEdit(channelId, messageId)
	emptyString := ""
	edit.Content = &emptyString
	edit.Embeds = &[]*discordgo.MessageEmbed{{Title: title, Description: description, Color: embedColor}}
	if components != nil {
		edit.Components = &[]discordgo.MessageComponent{discordgo.ActionsRow{Components: components}}
	}
	_, err := session.ChannelMessageEditComplex(edit)
	return err
}

func renderPosition(position, maxRaceLength, trackWidth float64) string {
	trail := int((math.Min(position, maxRaceLength)/maxRaceLength)*trackWidth) - 1
	line := strings.Repeat(".", int(math.Max(0.0, float64(trail))))
	line += "🐌"

	return fmt.Sprintf("%-20s", line)
}

func buildTrack(state snailrace.RaceState) string {
	description := "                          🏁\n"
	description += "  |-----------------------|\n"

	for index := range state.Snails {
		position := 100.0
		if state.Step < len(state.SnailPositions[index]) {
			position = state.SnailPositions[index][state.Step]
		}
		line := renderPosition(position, RaceLength, TrackWidth)

		if state.Step >= len(state.SnailPositions[index]) {
			if place, ok := state.Place[index]; ok {
				description += fmt.Sprintf("%2d| %s %d\n", index, line, place)
			} else {
				description += fmt.Sprintf("%2d| %s ?\n", index, line)
			}
		} else {
			description += fmt.Sprintf("%2d| %s |\n", index, line)
		}
	}
	description += "  |-----------------------|\n\n"
	return description
}
