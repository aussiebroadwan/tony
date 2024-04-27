package render

import (
	"fmt"

	"github.com/aussiebroadwan/tony/pkg/snailrace"
	"github.com/bwmarrin/discordgo"
)

var betPayout = map[int]func(snailrace.RaceState, snailrace.UserBet, map[int]int, func(string, int64)){
	snailrace.BetTypeWin:      payWinBet,
	snailrace.BetTypePlace:    payPlaceBet,
	snailrace.BetTypeEachWay:  payEachWayBet,
	snailrace.BetTypeQuinella: payQuinellaBet,
	snailrace.BetTypeExacta:   payExactaBet,
	snailrace.BetTypeTrifecta: payTrifectaBet,
}

func finishedMessage(state snailrace.RaceState, creditUser func(string, int64)) (string, []discordgo.MessageComponent) {

	description := fmt.Sprintf("```\nRace ID: %s\n\n%s\n", state.Race.Id, buildTrack(state))
	entrants := "Results:\n"

	// Display the results
	for index, snail := range state.Snails {
		entrants += fmt.Sprintf("[%d]: %s ", index, snail.Name)
		if place, ok := state.Place[index]; ok {
			switch place {
			case 1:
				entrants += "🥇"
			case 2:
				entrants += "🥈"
			case 3:
				entrants += "🥉"
			}
		}
		entrants += "\n"
	}
	description += entrants + "```"

	// Payout the winners
	for _, userBet := range state.Race.UserBets {
		if payout, ok := betPayout[userBet.Type]; ok {
			payout(state, userBet, state.Place, creditUser)
		}
	}

	return description, []discordgo.MessageComponent{
		discordgo.Button{
			Label:    "Concluded",
			Disabled: true,
			Style:    discordgo.SuccessButton,
			CustomID: "snailrace.host:" + state.Race.Id,
		},
	}
}

func payWinBet(state snailrace.RaceState, bet snailrace.UserBet, place map[int]int, creditUser func(string, int64)) {
	if place, ok := place[bet.Snail1Index]; ok {
		if place == 1 {
			odds := snailrace.CalculateOdds(state.Race.Pool, state.Race.Snails[bet.Snail1Index].Pool)
			returns := int64(float64(bet.Amount) * odds)
			creditUser(bet.UserId, returns)
		}
	}
}
func payPlaceBet(state snailrace.RaceState, bet snailrace.UserBet, place map[int]int, creditUser func(string, int64)) {
	if place, ok := place[bet.Snail1Index]; ok {
		if place <= 3 {
			odds := snailrace.CalculatePlaceOdds(state.Race.Pool, state.Race.Snails[bet.Snail1Index].Pool)
			returns := int64(float64(bet.Amount) * odds)
			creditUser(bet.UserId, returns)
		}
	}
}
func payEachWayBet(state snailrace.RaceState, bet snailrace.UserBet, place map[int]int, creditUser func(string, int64)) {
	if place, ok := place[bet.Snail1Index]; ok {

		if place <= 3 {
			// Calculate the place amount as half the bet amount is placed on
			// the win and the other half is on a place
			amount := float64(bet.Amount) / 2.0

			// Check if the snail placed
			placeOdds := snailrace.CalculatePlaceOdds(state.Race.Pool, state.Race.Snails[bet.Snail1Index].Pool)
			returns := int64(amount * placeOdds)

			// Check if the snail won
			if place == 1 {
				winOdds := snailrace.CalculateOdds(state.Race.Pool, state.Race.Snails[bet.Snail1Index].Pool)
				win := int64(amount * winOdds)
				returns += win
			}

			creditUser(bet.UserId, returns)
		}
	}
}
func payQuinellaBet(state snailrace.RaceState, bet snailrace.UserBet, place map[int]int, creditUser func(string, int64)) {
	snail1Place, ok1 := place[bet.Snail1Index]
	snail2Place, ok2 := place[bet.Snail2Index]
	if ok1 && ok2 {
		if (snail1Place == 1 && snail2Place == 2) || (snail1Place == 2 && snail2Place == 1) {
			odds := snailrace.CalculatePlaceOdds(state.Race.Pool, state.Race.Snails[bet.Snail1Index].Pool)
			odds *= snailrace.CalculatePlaceOdds(state.Race.Pool, state.Race.Snails[bet.Snail2Index].Pool)
			returns := int64(float64(bet.Amount) * odds)
			creditUser(bet.UserId, returns)
		}
	}
}
func payExactaBet(state snailrace.RaceState, bet snailrace.UserBet, place map[int]int, creditUser func(string, int64)) {
	snail1Place, ok1 := place[bet.Snail1Index]
	snail2Place, ok2 := place[bet.Snail2Index]
	if ok1 && ok2 {
		if snail1Place == 1 && snail2Place == 2 {
			odds := snailrace.CalculateOdds(state.Race.Pool, state.Race.Snails[bet.Snail1Index].Pool)
			odds *= snailrace.CalculatePlaceOdds(state.Race.Pool, state.Race.Snails[bet.Snail2Index].Pool)
			returns := int64(float64(bet.Amount) * odds)
			creditUser(bet.UserId, returns)
		}
	}
}
func payTrifectaBet(state snailrace.RaceState, bet snailrace.UserBet, place map[int]int, creditUser func(string, int64)) {
	snail1Place, ok1 := place[bet.Snail1Index]
	snail2Place, ok2 := place[bet.Snail2Index]
	snail3Place, ok3 := place[bet.Snail3Index]

	if ok1 && ok2 && ok3 {
		if snail1Place == 1 && snail2Place == 2 && snail3Place == 3 {
			odds := snailrace.CalculateOdds(state.Race.Pool, state.Race.Snails[bet.Snail1Index].Pool)
			odds *= snailrace.CalculatePlaceOdds(state.Race.Pool, state.Race.Snails[bet.Snail2Index].Pool)
			odds *= snailrace.CalculatePlaceOdds(state.Race.Pool, state.Race.Snails[bet.Snail3Index].Pool)
			returns := int64(float64(bet.Amount) * odds)
			creditUser(bet.UserId, returns)
		}
	}
}
