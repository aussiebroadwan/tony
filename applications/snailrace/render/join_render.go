package render

import (
	"fmt"
	"time"

	"github.com/aussiebroadwan/tony/pkg/snailrace"
	"github.com/bwmarrin/discordgo"
)

// joinMessage generates the join stage message and components.
func joinMessage(state snailrace.RaceState) (string, []discordgo.MessageComponent) {
	description := fmt.Sprintf(
		"A new race has been hosted!\n\nRace ID: `%s`\nStarting: `%s`\n\n"+
			"Click the `Join` button to join with your own snail.\n\n",
		state.Race.ID,
		state.Race.StartAt.Format(time.DateTime),
	)

	if len(state.Snails) == 0 {
		description += "> No snails have joined yet\n"
	} else {
		description += "**Entrants:**\n"
		for _, snail := range state.Snails {
			description += fmt.Sprintf("- %s <@%s>\n", snail.Name, snail.OwnerId)
		}
	}

	return description, []discordgo.MessageComponent{
		discordgo.Button{
			Label:    "Join",
			Style:    discordgo.SuccessButton,
			CustomID: "snailrace.host:join_request:" + state.Race.ID,
		},
	}
}
