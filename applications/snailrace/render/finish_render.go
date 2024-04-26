package render

import (
	"fmt"

	"github.com/aussiebroadwan/tony/pkg/snailrace"
	"github.com/bwmarrin/discordgo"
)

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
		if place, ok := state.Place[userBet.SnailIndex]; ok {
			if place == 1 {
				odds := snailrace.CalculateOdds(state.Race.Pool, state.Race.Snails[userBet.SnailIndex].Pool)
				win := int64(float64(userBet.Amount) * odds)
				creditUser(userBet.UserId, win)
			}
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
