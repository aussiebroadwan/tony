package blackjackApp

import (
	"github.com/aussiebroadwan/tony/framework"
	"github.com/aussiebroadwan/tony/pkg/blackjack"
	"github.com/bwmarrin/discordgo"
)

func stateRenderer(ctx framework.CommandContext) (func(blackjack.GameState, string, string), string, string) {

	session := ctx.Session()
	interaction := ctx.Interaction()

	// Post a new message to house the game
	msg, err := session.ChannelMessageSend(interaction.ChannelID, "Blackjack game in preparing")
	if err != nil {
		ctx.Logger().WithError(err).Error("Failed to create game message")
		return nil, "", ""
	}

	return func(state blackjack.GameState, channelId string, messageId string) {
		ctx.Logger().Info("Rendering game state")
		OnRender(session, state, channelId, messageId)
	}, interaction.ChannelID, msg.ID
}

func OnRender(session *discordgo.Session, state blackjack.GameState, channelId, messageId string) {
	JoinStateRender(session, state, channelId, messageId)
}

func JoinStateRender(session *discordgo.Session, state blackjack.GameState, channelId, messageId string) error {

	// Render the game state
	edit := discordgo.NewMessageEdit(channelId, messageId)
	edit.Embeds = &[]*discordgo.MessageEmbed{
		{
			Title:       "Blackjack: Joining Phase",
			Description: "To join place a bet. How much would you like to bet? Min is :coin: 10 and max is :coin: 999",
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
