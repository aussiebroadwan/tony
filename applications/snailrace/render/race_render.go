package render

import (
	"fmt"

	"github.com/aussiebroadwan/tony/pkg/snailrace"
	"github.com/bwmarrin/discordgo"
)

func raceMessage(state snailrace.RaceState) (string, []discordgo.MessageComponent) {

	description := fmt.Sprintf("```\nRace ID: %s\n\n%s\n", state.Race.Id, buildTrack(state))
	entrants := "Entrants:\n"

	for index, snail := range state.Snails {
		odds := snailrace.CalculateOdds(state.Race.Pool, state.Race.Snails[index].Pool)
		entrants += fmt.Sprintf("[%d]: %s @ %.02f\n", index, snail.Name, odds)
	}
	description += entrants + "```"

	return description, []discordgo.MessageComponent{
		discordgo.Button{
			Label:    "Racing",
			Disabled: true,
			Style:    discordgo.SuccessButton,
			CustomID: "snailrace.host:" + state.Race.Id,
		},
	}
}
