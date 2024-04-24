package snailrace_app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aussiebroadwan/tony/framework"
	"github.com/aussiebroadwan/tony/pkg/snailrace"
	"github.com/aussiebroadwan/tony/pkg/wallet"
	"github.com/bwmarrin/discordgo"
)

type SnailraceBetSubCommand struct {
	framework.ApplicationEvent
}

func (c SnailraceBetSubCommand) GetType() framework.AppType {
	return framework.AppTypeEvent
}

func (c SnailraceBetSubCommand) OnEvent(ctx framework.EventContext, eventType discordgo.InteractionType) {
	values := strings.Split(ctx.EventValue(), ":")

	if eventType != discordgo.InteractionMessageComponent {
		ctx.Logger().Error("Invalid event type")
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "**Error**: Invalid event type",
			},
		})
	}

	switch values[0] {
	case "win":
		handleWinBet(ctx, values[1], values[2])
	case "win_request":
		handleWinRequest(ctx, values[1])
	default:
		ctx.Logger().Error("Invalid event type")
	}

}

func handleWinRequest(ctx framework.EventContext, raceID string) {
	data := ctx.Interaction().MessageComponentData()
	snailIndex := data.Values[0]

	// You can react to button presses with no data and it doesn't error or send a message
	err := ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: fmt.Sprintf("snailrace.bet:win:%s:%s", raceID, snailIndex),
			Title:    "Quickbet: How much would you like to bet?",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "bet",
							Label:       "Bet Amount",
							Style:       discordgo.TextInputShort,
							Placeholder: "eg. 15",
							Required:    true,
							MaxLength:   3,
							MinLength:   2,
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

func handleWinBet(ctx framework.EventContext, raceId string, snailIndex string) {
	user := ctx.GetUser()
	snailIdx, _ := strconv.Atoi(snailIndex)

	// Fetch the user required data
	data := ctx.Interaction().ModalSubmitData()
	bet := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	betInt, err := strconv.Atoi(bet)
	if err != nil {
		ctx.Logger().WithError(err).Error("Failed to convert bet to integer")
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "**Error**: Invalid bet amount, must be an integer between 10 and 999",
			},
		})
		return
	}

	// Register the bet with the snailrace API
	err = snailrace.PlaceBet(user.ID, raceId, snailIdx, int64(betInt))
	if err != nil {
		ctx.Logger().WithError(err).Errorf("User %s has failed to place quickbet on race %s", user.Username, raceId)
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "**Error**: " + err.Error(),
			},
		})
		return
	}

	// Charge the user's balance
	err = wallet.Debit(ctx.Database(), user.ID, int64(betInt), "Snailrace Quickbet", "snailrace")
	if err != nil {
		// You can react to button presses with no data and it doesn't error or send a message
		ctx.Logger().WithError(err).Error("Failed to charge user")
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "**Error**: " + err.Error(),
			},
		})
		return
	}

	// You can react to button presses with no data and it doesn't error or send a message
	ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: "Bet placed",
		},
	})
}
