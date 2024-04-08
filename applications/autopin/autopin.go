package autopin

import "github.com/aussiebroadwan/tony/framework"

const autoPinThreshold = 1

func RegisterAutopinApp(bot *framework.Bot) framework.Route {
	return framework.NewRoute(bot, "autopin", &AutopinApp{})
}

// AutoPinRule is a rule that automatically pins messages that are reacted to
// with a pin emoji 📌 at least 5 times.
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
	count, pinned, err := GetAutopin(db, reaction.ChannelID, reaction.MessageID)
	if err != nil {
		return
	}

	// Check if the message should be pinned
	if count >= autoPinThreshold && !pinned.Valid {
		// Pin the message
		if err := ctx.Session().ChannelMessagePin(reaction.ChannelID, reaction.MessageID); err == nil {
			SetAutopinPinned(db, reaction.ChannelID, reaction.MessageID, true)
		}
	} else if count < autoPinThreshold && pinned.Valid {
		// Unpin the message
		if err := ctx.Session().ChannelMessageUnpin(reaction.ChannelID, reaction.MessageID); err != nil {
			SetAutopinPinned(db, reaction.ChannelID, reaction.MessageID, false)
		}
	}

}
