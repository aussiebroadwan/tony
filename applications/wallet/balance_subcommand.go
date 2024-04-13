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
	db := ctx.Database()

	user := ctx.GetUser()

	// Get the user's balance
	balance, err := wallet.Balance(db, user.ID)
	if err != nil {
		ctx.Logger().Errorf("Failed to get balance: %v", err)
		sendErrorResponse(ctx, "**Error:** Failed to get balance")
		return
	}

	// Get last 5 transactions
	transactions, err := wallet.History(db, user.ID, 5)
	if err != nil {
		ctx.Logger().Errorf("Failed to get transaction history: %v", err)
		sendErrorResponse(ctx, "**Error:** Failed to get transaction history")
		return
	}

	embed := createWalletBalanceEmbed(balance, transactions)
	sendEmbedResponse(ctx, embed)
}

// createWalletBalanceEmbed constructs a Discord message embed displaying the
// wallet balance and a summary of recent transactions.
func createWalletBalanceEmbed(balance int64, transactions []wallet.Transaction) *discordgo.MessageEmbed {
	body := formatTransactions(transactions)

	embed := &discordgo.MessageEmbed{
		Title:       "Wallet Balance",
		Description: fmt.Sprintf("Your current have :coin: %d in your wallet. ", balance),
		Color:       0x4CAF50,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Last 5 Transactions",
				Inline: false,
				Value:  fmt.Sprintf("```\n%s\n```", body),
			},
		},
	}

	return embed
}

// formatTransactions formats a list of transactions into a human-readable
// string. Example:
//
// ```
//
//	 -5 | Payment to uqcs-tony
//	-30 | Payment to uqcs-tony
//	 30 | Payment from lcox74
//	-30 | Payment to lcox74
//	-10 | Payment to uqcs-tony
//
// ```
func formatTransactions(transactions []wallet.Transaction) string {
	body := ""
	for _, transaction := range transactions {
		description := wordWrap(transaction.Description, 32, "      | ")
		amount := formatAmount(transaction.Amount, transaction.Type)
		body += fmt.Sprintf("%5s | %s\n", amount, description)
	}
	return body
}

// formatAmount formats a transaction amount into a string based on its type
// and value.
func formatAmount(amount int64, tType wallet.TransactionType) string {
	sign := ""
	if tType == wallet.DEBIT {
		sign = "-"
	}

	if amount > 1000 {
		return fmt.Sprintf("%s%.2fK", sign, float64(amount)/1000.0)
	}
	return fmt.Sprintf("%s%d", sign, amount)
}

// wordWrap breaks a long string into multiple lines at a specified width, with
// an optional prefix for each new line.
func wordWrap(text string, width int, prefix string) string {
	if len(text) <= width {
		return text
	}

	wrapped := text[:width] + "\n"
	for i := width; i < len(text); i += width {
		end := i + width
		if end > len(text) {
			end = len(text)
		}
		wrapped += prefix + text[i:end] + "\n"
	}
	return wrapped
}
