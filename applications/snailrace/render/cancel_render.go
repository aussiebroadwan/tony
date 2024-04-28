package render

import (
	"github.com/aussiebroadwan/tony/pkg/snailrace"
	"github.com/bwmarrin/discordgo"
)

func cancelledMessage(state snailrace.RaceState) (string, []discordgo.MessageComponent) {

	description := "Race has been cancelled due to not enough players.\n"

	return description, []discordgo.MessageComponent{
		discordgo.Button{
			Label:    "Concluded",
			Disabled: true,
			Style:    discordgo.SuccessButton,
			CustomID: "snailrace.host:" + state.Race.ID,
		},
	}
}
