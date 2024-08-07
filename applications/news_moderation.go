package applications

import (
	"errors"
	"strings"

	"github.com/aussiebroadwan/tony/framework"
	"github.com/bwmarrin/discordgo"
)

func RegisterNewsModeration(bot *framework.Bot) framework.Route {
	return framework.NewRoute(bot, "tech-news", &ModerateNewsRule{})
}

// The #tech-news channel is for news about technology. This rule will moderate
// the news in that channel. More specifically it will check and enforce the
// news post format.
//
// The format is as follows:
//
//	# <title>
//	{description}
//	<link>
//
// Example:
//
//	# CES is the perfect time to hide layoffs
//
//	CES just happened, if you were confused on why there are a ton of concepts...
//
//	https://www.theverge.com/2024/1/11/24034124/google-layoffs-engineering-assistant-hardware
//
// If the message does not match the format, the bot will delete the message and
// send a message to the user to let them know that the message was deleted and
// why.
type ModerateNewsRule struct {
	framework.ApplicationMessage
}

var (
	ErrInvalidNewsPostFormat = errors.New("news posts must be in the following format:\n```\n# <title>\n{description}\n<link>```")
	ErrTitleFormatError      = errors.New(`the title must start with '# '`)
	ErrLinkFormatError       = errors.New(`the link must be a web link (http or https)`)
)

func (r ModerateNewsRule) GetType() framework.AppType {
	return framework.AppTypeMessage
}

func (r ModerateNewsRule) OnMessage(ctx framework.MessageContext, channel *discordgo.Channel) {

	// Check if the message is in the correct channel
	if channel.Name != "tech-news" {
		return
	}

	content := ctx.Message().Content

	// Test the message content
	violation := r.test(content)
	if violation == nil {
		return
	}

	// Delete the message
	ctx.Session().ChannelMessageDelete(ctx.Message().ChannelID, ctx.Message().ID)
	ctx.Logger().Info("Deleted message: ", ctx.Message().ID)

	// Get or create a DM channel with the user
	dmChannel, err := ctx.Session().UserChannelCreate(ctx.Message().Author.ID)
	if err != nil {
		// Handle error, log it, or take appropriate action
		return
	}

	// Send a direct message to the user
	ctx.Session().ChannelMessageSend(dmChannel.ID, violation.Error())

	// Send a copy of the original content to the user
	sendOriginalMessageAsCodeblock(ctx, dmChannel.ID, content)
}

// Tests the rule against the content
func (r ModerateNewsRule) test(content string) error {
	// Split the message into lines
	lines := strings.Split(content, "\n")

	// Check if the message is in the correct format
	if len(lines) < 3 {
		return ErrInvalidNewsPostFormat
	}

	// Check if the title is in the correct format
	if !strings.HasPrefix(lines[0], "# ") {
		return ErrTitleFormatError
	}

	// Check if the link is in the correct format
	if !strings.HasPrefix(lines[len(lines)-1], "http") {
		return ErrLinkFormatError
	}

	return nil
}
