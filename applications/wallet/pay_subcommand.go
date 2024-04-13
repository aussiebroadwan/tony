package walletApp

import (
	"fmt"

	"github.com/aussiebroadwan/tony/framework"
	"github.com/aussiebroadwan/tony/pkg/wallet"
	"github.com/bwmarrin/discordgo"
)

type WalletPaySubCommand struct {
	framework.ApplicationSubCommand
}

func (c WalletPaySubCommand) GetType() framework.AppType {
	return framework.AppTypeSubCommand
}

func (c WalletPaySubCommand) OnCommand(ctx framework.CommandContext) {
	session := ctx.Session()
	interaction := ctx.Interaction()
	db := ctx.Database()

	user := interaction.User
	if user == nil {
		user = interaction.Member.User
	}

	commandOptions := interaction.ApplicationCommandData().Options[0].Options

	targetUserOpt, err := framework.GetOption(commandOptions, "user")
	if err != nil {
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "**Error:** User is required",
			},
		})
		return
	}
	targetUser := targetUserOpt.UserValue(session)

	amountOpt, err := framework.GetOption(commandOptions, "amount")
	if err != nil {
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "**Error:** Amount is required",
			},
		})
		return
	}
	amount := amountOpt.IntValue()
	if amount <= 0 {
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "**Error:** Amount must be greater than 0",
			},
		})
		return
	}

	// Debit the user's wallet
	err = wallet.Debit(db, user.ID, amount, fmt.Sprintf("Payment to %s", targetUser.Username), "wallet.pay")
	if err != nil {
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "**Error:** Failed to debit wallet",
			},
		})
		return
	}

	// Credit the target user's wallet
	err = wallet.Credit(db, targetUser.ID, amount, fmt.Sprintf("Payment from %s", user.Username), "wallet.pay")
	if err != nil {
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "**Error:** Failed to credit wallet",
			},
		})
		return
	}

	ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: "Payment successful",
		},
	})

	// Notify the target user
	dmChannel, err := ctx.Session().UserChannelCreate(targetUser.ID)
	if err != nil {
		// Handle error, log it, or take appropriate action
		return
	}

	// Send a direct message to the user
	ctx.Session().ChannelMessageSend(dmChannel.ID, fmt.Sprintf("You have received a payment of $%d from %s", amount, user.Mention()))
}
