package walletApp

import (
	"github.com/aussiebroadwan/tony/framework"
	"github.com/bwmarrin/discordgo"
)

func RegisterWalletApp(bot *framework.Bot) framework.Route {
	return framework.NewRoute(bot, "wallet",
		// Wallet
		&WalletAppCommand{}, // [NOP]

		// Subcommands
		framework.NewRoute(bot, "balance", &WalletBalanceSubCommand{}),
		framework.NewRoute(bot, "pay", &WalletPaySubCommand{}),
	)
}

type WalletAppCommand struct {
	framework.ApplicationMessage
}

func (c WalletAppCommand) GetType() framework.AppType {
	return framework.AppTypeCommand
}

func (c WalletAppCommand) GetDefinition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "wallet",
		Description: "Allows users to manage their wallet",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "balance",
				Description: "Check your wallet balance",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "pay",
				Description: "Pay another user",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "The user to pay",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "amount",
						Description: "The amount to pay",
						Required:    true,
					},
				},
			},
		},
	}
}

func (c WalletAppCommand) OnCommand(ctx framework.CommandContext) {
	// [NOP]
}
