package render

import (
	"fmt"
	"time"

	"github.com/aussiebroadwan/tony/pkg/snailrace"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// bettingMessage generates the message and componenets required for the
// betting stage of a race.
func bettingMessage(state snailrace.RaceState) (string, []discordgo.MessageComponent) {
	description := fmt.Sprintf(
		"Bets are now open to everyone, do you feel lucky? To place a quick bet you can select the snail via the drop down. \n\n```\nRace ID: %s\nStarting: %s\n\nEntrants:\n",
		state.Race.Id,
		state.Race.StartAt.Format(time.DateTime),
	)

	menuOptions := make([]discordgo.SelectMenuOption, len(state.Snails))
	for index, snail := range state.Snails {
		logrus.Infof("Calculating odds for snail %s, race_pool %d, pool %d", snail.Name, state.Race.Pool, state.Race.Snails[index].Pool)
		odds := snailrace.CalculateOdds(state.Race.Pool, state.Race.Snails[index].Pool)
		description += fmt.Sprintf("[%d]: %s @ %.02f\n", index, snail.Name, odds)

		menuOptions[index] = discordgo.SelectMenuOption{
			Label:   fmt.Sprintf("%s @ %.02f", snail.Name, odds),
			Value:   fmt.Sprintf("%d", index),
			Default: false,
		}
	}
	description += "```"

	return description, []discordgo.MessageComponent{
		discordgo.SelectMenu{
			CustomID:    "snailrace.bet:win_request:" + state.Race.Id,
			Placeholder: "Select a Quickbet",
			Options:     menuOptions,
		},
	}
}
