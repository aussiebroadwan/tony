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

		switch raceState.State {
		case snailrace.StateJoining:
			description, components = joinMessage(raceState)
		case snailrace.StateBetting:
			description, components = bettingMessage(raceState)
		case snailrace.StateInProgress:
			// description, components = progressMessage(raceState)
		case snailrace.StateFinished:
			description = "Under construction..."
		case snailrace.StateCancelled:
			description = "Under construction..."
		default:
			session.ChannelMessageEdit(channelId, messageId, preparingGameMessage)
			return
		}

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
		"A new race has been hosted!\n\nRace ID: `%s`\nStarting: `%s`\n\n"+
			"Click the `Join` button to join with your own snail.\n\n",
		state.Race.Id,
		state.Race.StartAt.Format(time.DateTime),
	)

	if len(state.Snails) == 0 {
		description += "> No snails have joined yet\n"
	} else {
		description += "**Entrants:**\n"
		for _, snail := range state.Snails {
			description += fmt.Sprintf("- %s <@%s>\n", snail.Name, snail.OwnerId) // TODO: Add owner mention
		}
	}

	return description, []discordgo.MessageComponent{
		discordgo.Button{
			Label:    "Join",
			Style:    discordgo.SuccessButton,
			CustomID: "snailrace.host:join_request:" + state.Race.Id,
		},
	}
}

// bettingMessage generates the message and componenets required for the
// betting stage of a race.
func bettingMessage(state snailrace.RaceState) (string, []discordgo.MessageComponent) {
	description := fmt.Sprintf(
		"Bets are now open to everyone, do you feel lucky? To place a quick bet you can select the snail via the drop down. \n\nRace ID: `%s`\nStarting: `%s`\n\n**Entrants:**\n",
		state.Race.Id,
		state.Race.StartAt.Format(time.DateTime),
	)

	menuOptions := make([]discordgo.SelectMenuOption, len(state.Snails))
	for index, snail := range state.Snails {
		logrus.Infof("Calculating odds for snail %s, race_pool %d, pool %d", snail.Name, state.Race.Pool, state.Race.Snails[index].Pool)
		odds := snailrace.CalculateOdds(state.Race.Pool, state.Race.Snails[index].Pool)
		description += fmt.Sprintf("`[%d]: %.02f` %s\n", index, odds, snail.Name)

		menuOptions[index] = discordgo.SelectMenuOption{
			Label:   fmt.Sprintf("%s @ %.02f", snail.Name, odds),
			Value:   fmt.Sprintf("%d", index),
			Default: false,
		}
	}

	return description, []discordgo.MessageComponent{
		discordgo.SelectMenu{
			CustomID:    "snailrace.bet:win_request:" + state.Race.Id,
			Placeholder: "Select a snail",
			Options:     menuOptions,
		},
	}
}
