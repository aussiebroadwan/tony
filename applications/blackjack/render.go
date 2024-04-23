package blackjack_app

import (
	"fmt"

	"github.com/aussiebroadwan/tony/framework"
	"github.com/aussiebroadwan/tony/pkg/blackjack"
	"github.com/aussiebroadwan/tony/pkg/tradingcards"
	"github.com/aussiebroadwan/tony/pkg/wallet"
	"github.com/bwmarrin/discordgo"
)

// Constants for embed colors and game messages.
const (
	embedColor           = 0x000000
	preparingGameMessage = "Blackjack game in preparing"
)

// stateRenderer sets up the messaging and functionality for rendering game states in blackjack.
func stateRenderer(ctx framework.CommandContext) (blackjack.StateChangeCallback, blackjack.AchievementCallback, string, string) {
	session := ctx.Session()
	interaction := ctx.Interaction()
	database := ctx.Database()

	msg, err := session.ChannelMessageSend(interaction.ChannelID, preparingGameMessage)
	if err != nil {
		ctx.Logger().WithError(err).Error("Failed to create game message")
		return nil, nil, "", ""
	}

	creditUser := func(userId string, amount int64) {
		if err := wallet.Credit(database, userId, amount, "Blackjack returns", "blackjack"); err != nil {
			ctx.Logger().WithError(err).Error("Failed to credit user")
		}
	}

	return createGameStateRenderFunc(ctx, session, creditUser), onAchievement(ctx), interaction.ChannelID, msg.ID
}

// onAchievement creates a function to handle achievement unlocks. It will
// assign a card to the user if they unlock an achievement. If it fails to
// assign the card it will return false, notifying the game to try again when
// the condition is met.
func onAchievement(ctx framework.CommandContext) blackjack.AchievementCallback {
	return func(userId string, achievement string) bool {
		ctx.Logger().WithField("user", userId).Info("Achievement unlocked: " + achievement)

		// Assign the card to the user
		err := tradingcards.AssignCard(ctx.Database(), userId, achievement)
		if err != nil {
			ctx.Logger().WithError(err).Error("Failed to assign achievement card")
			return false
		}

		session := ctx.Session()
		interaction := ctx.Interaction()
		card := Cards[achievement]

		// Notify the user of their new card
		session.ChannelMessageSend(
			interaction.ChannelID,
			fmt.Sprintf("**Achievement Unlocked**: %s \n<@%s>",
				card.Title,
				userId,
			),
		)

		return true
	}
}

// createGameStateRenderFunc creates a function to render the game state based on the current stage.
func createGameStateRenderFunc(ctx framework.CommandContext, session *discordgo.Session, creditUser func(string, int64)) blackjack.StateChangeCallback {
	return func(stage blackjack.GameStage, state blackjack.GameState, channelId string, messageId string) {
		ctx.Logger().WithField("stage", stage).Info("Rendering game state")

		var err error
		description := ""
		var components []discordgo.MessageComponent = nil

		switch stage {
		case blackjack.JoinStage:
			description, components = joinMessage(state)
		case blackjack.RoundStage:
			description, components = roundMessage(state)
		case blackjack.PayoutStage:
			description, components = payoutMessage(state, creditUser)
		case blackjack.ReshuffleStage:
			description, components = reshuffleMessage()
		case blackjack.FinishedStage:
			description, components = finishedMessage()
		default:
			session.ChannelMessageEdit(channelId, messageId, preparingGameMessage)
			return
		}

		err = renderState(session, channelId, messageId, "Blackjack: "+string(stage), description, components)

		if err != nil {
			ctx.Logger().WithField("stage", stage).WithError(err).Error("Failed to render game state")
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
func joinMessage(state blackjack.GameState) (string, []discordgo.MessageComponent) {
	description := "To join place a bet. How much would you like to bet? Min is :coin: 10 and max is :coin: 999"
	if len(state.Users) > 0 {
		description += fmt.Sprintf("\n\nPlayers (%d / %d):\n", len(state.Users), blackjack.MaxPlayers)
		for _, user := range state.Users {
			description += fmt.Sprintf("<@%s> bets :coin: %d\n", user.Id, user.Bet)
		}
	}
	return description, []discordgo.MessageComponent{
		discordgo.Button{
			Label:    "Join",
			Style:    discordgo.SuccessButton,
			CustomID: "blackjack:host",
		},
	}
}

// roundMessage generates the round stage message and components.
func roundMessage(state blackjack.GameState) (string, []discordgo.MessageComponent) {
	description := ""
	if state.PlayerTurn < 0 {
		description = "The round is in progress, dealing cards now...\n\n"
	} else if state.PlayerTurn < len(state.Users) {
		playersTurn := state.Users[state.PlayerTurn].Id
		description = fmt.Sprintf("The round is in progress and its currently <@%s>'s turn to play.\n\n", playersTurn)
	} else {
		description = "The round is in progress, everyone's had their turn. Time for the dealer to play.\n\n"
	}

	// Build the dealer's hand
	description += fmt.Sprintf("Dealer (%d): ", state.Hand.Score())
	for _, card := range state.Hand {
		description += fmt.Sprintf("`%s%s` ", card.Rank, card.Suit)
	}
	if state.Hand.Score() > blackjack.MaximumHandScore {
		description += " - Bust"
	} else if state.Hand.Score() == blackjack.MaximumHandScore && len(state.Hand) == 2 {
		description += " - Blackjack"
	}
	description += "\n\n"

	// Build the board
	for _, user := range state.Users {
		description += fmt.Sprintf("<@%s> (%d): ", user.Id, user.Hand.Score())
		for _, card := range user.Hand {
			description += fmt.Sprintf("`%s%s` ", card.Rank, card.Suit)
		}

		if user.Hand.Score() > blackjack.MaximumHandScore {
			description += " - Bust"
		} else if user.Hand.Score() == blackjack.MaximumHandScore && len(user.Hand) == 2 {
			description += " - Blackjack"
		}

		description += "\n"
	}

	return description, []discordgo.MessageComponent{
		discordgo.Button{
			Label:    "Hit",
			Style:    discordgo.SuccessButton,
			CustomID: "blackjack:hit",
		},
		discordgo.Button{
			Label:    "Stand",
			Style:    discordgo.DangerButton,
			CustomID: "blackjack:stand",
		},
	}
}

// payoutMessage generates the payout stage message and components.
func payoutMessage(state blackjack.GameState, creditUser func(string, int64)) (string, []discordgo.MessageComponent) {
	description := "The round is over. Here are the results:\n\n"
	for _, user := range state.Users {
		description += fmt.Sprintf("<@%s>: :coin: %d\n", user.Id, user.Bet)
		if user.Bet > 0 {
			creditUser(user.Id, user.Bet)
		}
	}
	description += "\nThe next round will begin shortly."
	return description, []discordgo.MessageComponent{
		discordgo.Button{
			Label:    "Hit",
			Style:    discordgo.SuccessButton,
			CustomID: "blackjack:hit",
			Disabled: true,
		},
		discordgo.Button{
			Label:    "Stand",
			Style:    discordgo.DangerButton,
			CustomID: "blackjack:stand",
			Disabled: true,
		},
	} // No components needed for payout state
}

func reshuffleMessage() (string, []discordgo.MessageComponent) {
	return "The deck is being reshuffled. A new round will begin shortly...", []discordgo.MessageComponent{
		discordgo.Button{
			Label:    "Join",
			Style:    discordgo.SuccessButton,
			CustomID: "blackjack:host",
			Disabled: true,
		},
	}
}

func finishedMessage() (string, []discordgo.MessageComponent) {
	return "The game has finished. Start a new game with `/blackjack`.", []discordgo.MessageComponent{
		discordgo.Button{
			Label:    "Join",
			Style:    discordgo.SuccessButton,
			CustomID: "blackjack:host",
			Disabled: true,
		},
	}
}
