package snailrace_app

import (
	"os"

	"github.com/aussiebroadwan/tony/applications/snailrace/render"
	"github.com/aussiebroadwan/tony/framework"
	"github.com/aussiebroadwan/tony/pkg/snailrace"
	"github.com/bwmarrin/discordgo"
)

const CommonSnailUsage = 25

func RegisterSnailraceApp(bot *framework.Bot) framework.Route {
	return framework.NewRoute(bot, "snailrace",
		// snailrace
		&Snailrace{}, // [NOP]

		framework.NewRoute(bot, "host", &SnailraceHostSubCommand{}),
		framework.NewRoute(bot, "bet",
			&SnailraceBetSubCommand{}, // Handle Bet Events

			framework.NewRoute(bot, "win", &SnailraceBetSubCommand{}),
			framework.NewRoute(bot, "place", &SnailraceBetSubCommand{}),
			framework.NewRoute(bot, "eachway", &SnailraceBetSubCommand{}),
			framework.NewRoute(bot, "quinella", &SnailraceBetSubCommand{}),
			framework.NewRoute(bot, "exacta", &SnailraceBetSubCommand{}),
			framework.NewRoute(bot, "trifecta", &SnailraceBetSubCommand{}),
		),
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

	serverID := os.Getenv("DISCORD_SERVER_ID")
	channelName := os.Getenv("SNAILRACE_TV_CHANNEL")

	// Start the snailrace TV if the channel is set
	if channelID := framework.ChannelNameToID(ctx, serverID, channelName); channelID != "" {
		go snailrace.LaunchSnailraceTV(func() (snailrace.StateChangeCallback, snailrace.AchievementCallback, string, string) {
			return render.StateRenderer(ctx, channelID)
		})
	}
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
			{
				Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
				Name:        "bet",
				Description: "Bet on a race",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Name:        "win",
						Description: "Bet on a snail to win",
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionString,
								Name:        "race_id",
								Description: "The ID of the race to place a bet on",
								Required:    true,
							},
							{
								Type:        discordgo.ApplicationCommandOptionInteger,
								Name:        "bet_amount",
								Description: "The amount to bet",
								Required:    true,
							},
							{
								Type:        discordgo.ApplicationCommandOptionInteger,
								Name:        "snail_index_1",
								Description: "The index of the snail to win",
								Required:    true,
							},
						},
					},
					{
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Name:        "place",
						Description: "Bet on a snail to place in the top 3",
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionString,
								Name:        "race_id",
								Description: "The ID of the race to place a bet on",
								Required:    true,
							},
							{
								Type:        discordgo.ApplicationCommandOptionInteger,
								Name:        "bet_amount",
								Description: "The amount to bet",
								Required:    true,
							},
							{
								Type:        discordgo.ApplicationCommandOptionInteger,
								Name:        "snail_index_1",
								Description: "The index of the snail to place",
								Required:    true,
							},
						},
					},
					{
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Name:        "eachway",
						Description: "Bet on a snail to place in the top 3 and/or win",
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionString,
								Name:        "race_id",
								Description: "The ID of the race to place a bet on",
								Required:    true,
							},
							{
								Type:        discordgo.ApplicationCommandOptionInteger,
								Name:        "bet_amount",
								Description: "The amount to bet",
								Required:    true,
							},
							{
								Type:        discordgo.ApplicationCommandOptionInteger,
								Name:        "snail_index_1",
								Description: "The index of the snail to win/place",
								Required:    true,
							},
						},
					},
					{
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Name:        "quinella",
						Description: "Bet on two snails to win and place 2nd in any order",
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionString,
								Name:        "race_id",
								Description: "The ID of the race to place a bet on",
								Required:    true,
							},
							{
								Type:        discordgo.ApplicationCommandOptionInteger,
								Name:        "bet_amount",
								Description: "The amount to bet",
								Required:    true,
							},
							{
								Type:        discordgo.ApplicationCommandOptionInteger,
								Name:        "snail_index_1",
								Description: "The index of the snail to win or place 2nd",
								Required:    true,
							},
							{
								Type:        discordgo.ApplicationCommandOptionInteger,
								Name:        "snail_index_2",
								Description: "The index of the snail to win or place 2nd",
								Required:    true,
							},
						},
					},
					{
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Name:        "exacta",
						Description: "Bet on two snails to win and place 2nd in order",
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionString,
								Name:        "race_id",
								Description: "The ID of the race to place a bet on",
								Required:    true,
							},
							{
								Type:        discordgo.ApplicationCommandOptionInteger,
								Name:        "bet_amount",
								Description: "The amount to bet",
								Required:    true,
							},
							{
								Type:        discordgo.ApplicationCommandOptionInteger,
								Name:        "snail_index_1",
								Description: "The index of the snail to win",
								Required:    true,
							},
							{
								Type:        discordgo.ApplicationCommandOptionInteger,
								Name:        "snail_index_2",
								Description: "The index of the snail to place 2nd",
								Required:    true,
							},
						},
					},
					{
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Name:        "trifecta",
						Description: "Bet on three snails to come in the top 3 in order",
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionString,
								Name:        "race_id",
								Description: "The ID of the race to place a bet on",
								Required:    true,
							},
							{
								Type:        discordgo.ApplicationCommandOptionInteger,
								Name:        "bet_amount",
								Description: "The amount to bet",
								Required:    true,
							},
							{
								Type:        discordgo.ApplicationCommandOptionInteger,
								Name:        "snail_index_1",
								Description: "The index of the snail to win",
								Required:    true,
							},
							{
								Type:        discordgo.ApplicationCommandOptionInteger,
								Name:        "snail_index_2",
								Description: "The index of the snail to place 2nd",
								Required:    true,
							},
							{
								Type:        discordgo.ApplicationCommandOptionInteger,
								Name:        "snail_index_3",
								Description: "The index of the snail to place 3rd",
								Required:    true,
							},
						},
					},
				},
			},
		},
	}
}

func (s Snailrace) OnCommand(ctx framework.CommandContext) {}
