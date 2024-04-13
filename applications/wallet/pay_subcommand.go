package walletApp

import (
	"fmt"

	"github.com/aussiebroadwan/tony/framework"
	"github.com/aussiebroadwan/tony/pkg/wallet"
	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

type WalletPaySubCommand struct {
	framework.ApplicationSubCommand
}

func (c WalletPaySubCommand) GetType() framework.AppType {
	return framework.AppTypeSubCommand
}

func (c WalletPaySubCommand) OnCommand(ctx framework.CommandContext) {
	session := ctx.Session()
	db := ctx.Database()

	user := ctx.GetUser()
	targetUser := ctx.GetOption("user").UserValue(session)
	amount := int64(ctx.GetOption("amount").IntValue())

	if !validateAmount(ctx, amount) {
		ctx.Logger().Error("Invalid amount")
		return
	}

	if err := processPayment(db, user, targetUser, amount); err != nil {
		ctx.Logger().Errorf("Failed to process payment: %v", err)
		sendErrorResponse(ctx, "**Error:** "+err.Error())
		return
	}

	sendSuccessResponse(ctx, "Payment successful")
	notifyTargetUser(ctx, targetUser, amount, user)
}

// validateAmount checks if the provided amount is a valid transaction amount.
func validateAmount(ctx framework.CommandContext, amount int64) bool {
	if amount <= 0 {
		sendErrorResponse(ctx, "**Error:** Amount must be greater than 0")
		return false
	}
	return true
}

// processPayment handles the transaction logic, including database operations
// to transfer funds.
func processPayment(db *gorm.DB, user *discordgo.User, targetUser *discordgo.User, amount int64) error {
	err := wallet.Trasfer(db,
		user.ID, targetUser.ID,
		amount,
		fmt.Sprintf("Payment to %s", targetUser.Username),
		fmt.Sprintf("Payment from %s", user.Username),
		"wallet.pay",
	)

	if err != nil {
		return fmt.Errorf("failed to process payment")
	}
	return nil
}

// notifyTargetUser sends a direct message to the recipient to notify them of
// the received payment.
func notifyTargetUser(ctx framework.CommandContext, targetUser *discordgo.User, amount int64, sender *discordgo.User) {
	dmChannel, err := ctx.Session().UserChannelCreate(targetUser.ID)
	if err != nil {
		ctx.Logger().Errorf("Failed to create DM channel with user %s", targetUser.ID)
		return
	}
	ctx.Session().ChannelMessageSend(dmChannel.ID, fmt.Sprintf("You have received a payment of :coin: $%d from %s", amount, sender.Username))
}
