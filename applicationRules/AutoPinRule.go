package applicationrules

import (
	"fmt"

	"github.com/aussiebroadwan/tony/database"
	"github.com/aussiebroadwan/tony/framework"
)

const (
	autoPinThreshold = 5
)

// AutoPinRule is a rule that automatically pins messages that are reacted to
// with a pin emoji 📌 at least 5 times.
type AutoPinRule struct {
	framework.ApplicationRule
}

func (r *AutoPinRule) Name() string {
	return "auto-pin"
}

func (r *AutoPinRule) GetType() framework.ApplicationRuleType {
	return framework.ApplicationRuleTypeReactions
}

// Test tests the rule against the content
func (r *AutoPinRule) Test(content string) error {
	if content != "📌" {
		return nil
	}
	return fmt.Errorf("pin emoji found in message")
}

// Action takes action if the rule is violated
func (r *AutoPinRule) Action(ctx *framework.Context, violation error) {
	db := ctx.Database()
	reaction, add := ctx.Reaction()

	// Increment or decrement the count
	if add {
		database.IncrementAutoPin(db, reaction.ChannelID, reaction.MessageID)
	} else {
		database.DecrementAutoPin(db, reaction.ChannelID, reaction.MessageID)
	}

	// Get the count
	count, pinned, err := database.GetAutoPin(db, reaction.ChannelID, reaction.MessageID)
	if err != nil {
		return
	}

	// Check if the message should be pinned
	if count >= autoPinThreshold && !pinned.Valid {
		// Pin the message
		if err := ctx.Session().ChannelMessagePin(reaction.ChannelID, reaction.MessageID); err == nil {
			database.SetAutoPinPinned(db, reaction.ChannelID, reaction.MessageID, true)
		}
	} else if count < autoPinThreshold && pinned.Valid {
		// Unpin the message
		if err := ctx.Session().ChannelMessageUnpin(reaction.ChannelID, reaction.MessageID); err != nil {
			database.SetAutoPinPinned(db, reaction.ChannelID, reaction.MessageID, false)
		}
	}
}
