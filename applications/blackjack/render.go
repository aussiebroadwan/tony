package blackjackApp

import (
	"fmt"

	"github.com/aussiebroadwan/tony/framework"
	"github.com/aussiebroadwan/tony/pkg/blackjack"
	"github.com/aussiebroadwan/tony/pkg/wallet"
	"github.com/bwmarrin/discordgo"
)

func stateRenderer(ctx framework.CommandContext) (func(blackjack.GameStage, blackjack.GameState, string, string), string, string) {

	session := ctx.Session()
	interaction := ctx.Interaction()
	database := ctx.Database()

	// Post a new message to house the game
	msg, err := session.ChannelMessageSend(interaction.ChannelID, "Blackjack game in preparing")
	if err != nil {
		ctx.Logger().WithError(err).Error("Failed to create game message")
		return nil, "", ""
	}

	creditUser := func(userId string, amount int64) {
		err := wallet.Credit(database, userId, amount, "Blackjack returns", "blackjack")
		if err != nil {
			ctx.Logger().WithError(err).Error("Failed to credit user")
		}
	}

	return func(stage blackjack.GameStage, state blackjack.GameState, channelId string, messageId string) {
		ctx.Logger().WithField("stage", stage).Info("Rendering game state")

		var err error

		switch stage {
		case blackjack.JoinStage:
			err = JoinStateRender(session, state, channelId, messageId)
		case blackjack.RoundStage:
			err = RoundStateRender(session, state, channelId, messageId)
		case blackjack.PayoutStage:
			err = PayoutStateRender(session, state, channelId, messageId, creditUser)
		default:
			_, err = session.ChannelMessageEdit(channelId, messageId, "Blackjack game in preparing")
		}

		if err != nil {
			ctx.Logger().WithField("stage", stage).WithError(err).Error("Failed to render game state")
		}

	}, interaction.ChannelID, msg.ID
}

func JoinStateRender(session *discordgo.Session, state blackjack.GameState, channelId, messageId string) error {

	// Build the game description
	description := "To join place a bet. How much would you like to bet? Min is :coin: 10 and max is :coin: 999"
	if len(state.Users) > 0 {
		description += fmt.Sprintf("\n\nPlayers (%d / %d):\n", len(state.Users), blackjack.MaxPlayers)
		for _, user := range state.Users {
			description += fmt.Sprintf("<@%s> bets %d\n", user.Id, user.Bet)
		}
	}

	// Render the game state
	edit := discordgo.NewMessageEdit(channelId, messageId)
	edit.Embeds = &[]*discordgo.MessageEmbed{
		{
			Title:       "Blackjack: Join",
			Description: description,
			Color:       0x2ecc71,
		},
	}
	edit.Components = &[]discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Join",
					Style:    discordgo.SuccessButton,
					CustomID: "blackjack:host",
				},
			},
		},
	}
	_, err := session.ChannelMessageEditComplex(edit)
	if err != nil {
		return err
	}

	return nil
}

func RoundStateRender(session *discordgo.Session, state blackjack.GameState, channelId, messageId string) error {
	description := ""
	if state.PlayerTurn < len(state.Users) {
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

	// Render the game state
	edit := discordgo.NewMessageEdit(channelId, messageId)
	edit.Embeds = &[]*discordgo.MessageEmbed{
		{
			Title:       "Blackjack: Playing",
			Description: description,
			Color:       0x2ecc71,
		},
	}
	edit.Components = &[]discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
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
			},
		},
	}
	_, err := session.ChannelMessageEditComplex(edit)
	if err != nil {
		return err
	}

	return nil
}

func PayoutStateRender(session *discordgo.Session, state blackjack.GameState, channelId, messageId string, creditUser func(userId string, amount int64)) error {

	description := "The round is over. Here are the results:\n\n"
	for _, user := range state.Users {
		description += fmt.Sprintf("<@%s> (%d): :coin: %d\n", user.Id, user.Hand.Score(), user.Bet)

		if user.Bet == 0 {
			continue
		}

		// Bets have been precalculated in the game state, so we can just credit the user directly
		creditUser(user.Id, user.Bet)
	}
	description += "\n\nThe next round will begin shortly."

	// Render the game state
	edit := discordgo.NewMessageEdit(channelId, messageId)
	edit.Embeds = &[]*discordgo.MessageEmbed{
		{
			Title:       "Blackjack: Payout",
			Description: description,
			Color:       0x2ecc71,
		},
	}

	return nil
}
