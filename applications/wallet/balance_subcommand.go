package walletApp

import (
	"fmt"

	"github.com/aussiebroadwan/tony/framework"
	"github.com/aussiebroadwan/tony/pkg/wallet"
	"github.com/bwmarrin/discordgo"
)

type WalletBalanceSubCommand struct {
	framework.ApplicationSubCommand
}

func (c WalletBalanceSubCommand) GetType() framework.AppType {
	return framework.AppTypeSubCommand
}

func (c WalletBalanceSubCommand) OnCommand(ctx framework.CommandContext) {
	interaction := ctx.Interaction()
	db := ctx.Database()

	user := interaction.User
	if user == nil {
		user = interaction.Member.User
	}

	// Get the user's balance
	balance, err := wallet.Balance(db, user.ID)
	if err != nil {
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "**Error:** Failed to get balance",
			},
		})
		return
	}

	// Get last 5 transactions
	transactions, err := wallet.History(db, user.ID, 5)
	if err != nil {
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "**Error:** Failed to get transaction history",
			},
		})
		return
	}

	embed := createWalletBalanceEmbed(balance, transactions)
	ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func createWalletBalanceEmbed(balance int64, transactions []wallet.Transaction) *discordgo.MessageEmbed {
	body := ""
	for _, transaction := range transactions {
		if transaction.Type == wallet.DEBIT {
			transaction.Amount = -transaction.Amount
		}

		// Word wrap the description to 32 characters
		wrappedDescription := ""
		if len(transaction.Description) > 32 {
			wrappedDescription = transaction.Description[:32] + "\n"

			// Add the rest of the description
			for i := 32; i < len(transaction.Description); i += 32 {
				end := i + 32
				if end > len(transaction.Description) {
					end = len(transaction.Description)
				}
				wrappedDescription += "      | " + transaction.Description[i:end] + "\n"
			}
		} else {
			wrappedDescription = transaction.Description + "\n"
		}

		// Add the transaction to the body
		body += fmt.Sprintf("%5d | %s",
			transaction.Amount,
			wrappedDescription,
		)

	}

	embed := &discordgo.MessageEmbed{
		Title:       "Wallet Balance",
		Description: "Your current wallet balance",
		Color:       0x4CAF50,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Balance",
				Value:  fmt.Sprintf("%d", balance),
				Inline: true,
			},
			{
				Name:   "Last 5 Transactions",
				Inline: false,
				Value:  fmt.Sprintf("```\n%s\n```", body),
			},
		},
	}

	return embed
}
