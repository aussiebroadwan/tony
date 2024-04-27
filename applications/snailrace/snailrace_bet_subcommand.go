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
	framework.ApplicationSubCommand
	framework.ApplicationEvent
}

func (c SnailraceBetSubCommand) GetType() framework.AppType {
	return framework.AppTypeEvent | framework.AppTypeSubCommand
}

var betConversion = map[string]int{
	"snailrace.bet.win":      snailrace.BetTypeWin,
	"snailrace.bet.place":    snailrace.BetTypePlace,
	"snailrace.bet.eachway":  snailrace.BetTypeEachWay,
	"snailrace.bet.quinella": snailrace.BetTypeQuinella,
	"snailrace.bet.exacta":   snailrace.BetTypeExacta,
	"snailrace.bet.trifecta": snailrace.BetTypeTrifecta,
}

func (c SnailraceBetSubCommand) OnCommand(ctx framework.CommandContext) {
	interaction := ctx.Interaction()
	route := ctx.GetRoute()
	user := ctx.GetUser()
	commandOptions := interaction.ApplicationCommandData().Options[0].Options

	// Required fields validation and error response handling
	requiredFields := []string{"race_id", "bet_amount", "snail_index_1"}
	optionsMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
	for _, fieldName := range requiredFields {
		option, err := framework.GetOption(commandOptions, fieldName)
		if err != nil {
			logAndRespondWithError(ctx, fmt.Sprintf("%s is required", fieldName), err)
			return
		}
		optionsMap[fieldName] = option
	}

	snailIdxs := []int{int(optionsMap["snail_index_1"].IntValue())}
	handleBetTypes(ctx, route, &snailIdxs, commandOptions)

	if err := snailrace.PlaceBet(user.ID, optionsMap["race_id"].StringValue(), optionsMap["bet_amount"].IntValue(), betConversion[route], snailIdxs...); err != nil {
		logAndRespondWithError(ctx, "Failed to place a bet", err)
		return
	}

	if err := wallet.Debit(ctx.Database(), user.ID, optionsMap["bet_amount"].IntValue(), "Snailrace Bet", "snailrace"); err != nil {
		logAndRespondWithError(ctx, "Failed to charge user", err)
		return
	}

	respond(ctx, "Bet placed")

}

// handleBetTypes processes additional bet types based on the command route. It
// modifies the list of snail indices based on the bet type (e.g., quinella,
// exacta, trifecta).
func handleBetTypes(ctx framework.CommandContext, route string, snailIdxs *[]int, options []*discordgo.ApplicationCommandInteractionDataOption) {
	if route == "snailrace.bet.quinella" || route == "snailrace.bet.exacta" {
		snailIndex2, err := framework.GetOption(options, "snail_index_2")
		if err != nil {
			logAndRespondWithError(ctx, "Snail Index 2 is required", err)
			return
		}
		*snailIdxs = append(*snailIdxs, int(snailIndex2.IntValue()))
	}
	if route == "snailrace.bet.trifecta" {
		handleTrifectaBet(ctx, snailIdxs, options)
	}
}

// handleTrifectaBet specifically handles the trifecta bet type, where three
// snail indices are required. It appends the indices to the snailIdxs slice
// after validating them.
func handleTrifectaBet(ctx framework.CommandContext, snailIdxs *[]int, options []*discordgo.ApplicationCommandInteractionDataOption) {
	requiredIndexes := []string{"snail_index_2", "snail_index_3"}
	for _, idx := range requiredIndexes {
		option, err := framework.GetOption(options, idx)
		if err != nil {
			logAndRespondWithError(ctx, fmt.Sprintf("%s is required", idx), err)
			return
		}
		*snailIdxs = append(*snailIdxs, int(option.IntValue()))
	}
}

// OnEvent handles events triggered within the context of snail racing
// quickbets. It will post the quickbet modal and handle the user's response.
func (c SnailraceBetSubCommand) OnEvent(ctx framework.EventContext, eventType discordgo.InteractionType) {
	values := strings.Split(ctx.EventValue(), ":")

	switch values[0] {
	case "win":
		if len(values) < 3 {
			logAndRespondWithError(ctx, "Incomplete event data for quickbet", fmt.Errorf("incomplete event data"))
			return
		}
		handleQuickBet(ctx, values[1], values[2])
	case "win_request":
		if len(values) < 2 {
			logAndRespondWithError(ctx, "Incomplete event data for quickbet request", fmt.Errorf("incomplete event data"))
			return
		}
		handleQuickbetRequest(ctx, values[1])
	default:
		logAndRespondWithError(ctx, "Invalid event type", fmt.Errorf("invalid event type"))
	}
}

// handleQuickbetRequest deals with a user's request to place a quicbet (win).
// It prompts the user to input the bet amount through a modal interaction.
func handleQuickbetRequest(ctx framework.EventContext, raceID string) {
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
		logAndRespondWithError(ctx, "Failed to respond to interaction", err)
	}
}

// handleQuickBet processes a win bet by validating the bet amount and snail
// index, placing the bet, and debiting the user's account if the bet is
// successfully placed.
func handleQuickBet(ctx framework.EventContext, raceId string, snailIndex string) {
	user := ctx.GetUser()
	snailIdx, err := strconv.Atoi(snailIndex)
	if err != nil {
		logAndRespondWithError(ctx, "Failed to convert snail index to integer", err)
		return
	}

	// Fetch the user required data from modal submit
	data := ctx.Interaction().ModalSubmitData()
	bet, err := strconv.Atoi(data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value)
	if err != nil {
		logAndRespondWithError(ctx, "Failed to convert bet to integer", err)
		return
	}

	// Register the bet with the snailrace API
	err = snailrace.PlaceBet(user.ID, raceId, int64(bet), snailrace.BetTypeWin, snailIdx)
	if err != nil {
		logAndRespondWithError(ctx, fmt.Sprintf("User %s has failed to place quickbet on race %s", user.Username, raceId), err)
		return
	}

	// Charge the user's balance
	err = wallet.Debit(ctx.Database(), user.ID, int64(bet), "Snailrace Quickbet", "snailrace")
	if err != nil {
		logAndRespondWithError(ctx, "Failed to charge user", err)
		return
	}

	respond(ctx, "Bet placed")
}

func logAndRespondWithError(ctx framework.ResponseContext, logMessage string, err error) {
	ctx.Logger().WithError(err).Error(logMessage)
	respond(ctx, fmt.Sprintf("**Error**: %s", err.Error()))
}

func respond(ctx framework.ResponseContext, message string) {
	ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: message,
		},
	})
}
