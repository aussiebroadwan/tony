package autopin

import "github.com/aussiebroadwan/tony/framework"

// AutoPinRule is a rule that automatically pins messages that are reacted to
// with a pin emoji 📌 at least x times.
type AutopinApp struct {
	framework.ApplicationReaction
}

func (a AutopinApp) GetType() framework.AppType {
	return framework.AppTypeReaction | framework.AppTypeMountable
}

func (a AutopinApp) OnMount(ctx framework.MountContext) {
	SetupAutopinDB(ctx.Database())
}

func (a AutopinApp) OnReaction(ctx framework.ReactionContext) {
	db := ctx.Database()
	reaction, add := ctx.Reaction()

	if reaction.Emoji.Name != "📌" {
		return
	}

	// Increment or decrement the count
	if add {
		IncrementAutopin(db, reaction.ChannelID, reaction.MessageID)
	} else {
		DecrementAutopin(db, reaction.ChannelID, reaction.MessageID)
	}

	// Get the count
	count, pinned, _ := GetAutopin(db, reaction.ChannelID, reaction.MessageID)

	// Check if the message should be pinned
	if count >= autopinThreshold && pinned == nil {
		// Pin the message
		if err := ctx.Session().ChannelMessagePin(reaction.ChannelID, reaction.MessageID); err == nil {
			SetAutopinPinned(db, reaction.ChannelID, reaction.MessageID, true)
		}
	} else if count < autopinThreshold && pinned != nil {
		// Unpin the message
		if err := ctx.Session().ChannelMessageUnpin(reaction.ChannelID, reaction.MessageID); err != nil {
			SetAutopinPinned(db, reaction.ChannelID, reaction.MessageID, false)
		}
	}

}
