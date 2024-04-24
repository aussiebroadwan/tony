package snailrace_app

import (
	"github.com/aussiebroadwan/tony/framework"
	"github.com/aussiebroadwan/tony/pkg/snailrace"
	"github.com/bwmarrin/discordgo"
)

func RegisterSnailraceApp(bot *framework.Bot) framework.Route {
	return framework.NewRoute(bot, "snailrace",
		// snailrace
		&Snailrace{}, // [NOP]

		framework.NewRoute(bot, "host", &SnailraceHostSubCommand{}),
	)
}

type Snailrace struct {
	framework.ApplicationCommand
	framework.ApplicationMountable
}

func (s Snailrace) GetType() framework.AppType {
	return framework.AppTypeCommand | framework.AppTypeMountable
}

func (s Snailrace) OnMount(ctx framework.MountContext) {
	snailrace.SetupSnailraceDB(ctx.Database())
}

func (s Snailrace) GetDefinition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "snailrace",
		Description: "Let's race snails!",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "host",
				Description: "Host a snailrace",
			},
		},
	}
}

func (s Snailrace) OnCommand(ctx framework.CommandContext) {}
