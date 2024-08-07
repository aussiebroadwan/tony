package applications

import (
	"strings"

	"github.com/aussiebroadwan/tony/framework"
)

func sendOriginalMessageAsCodeblock(ctx framework.MessageContext, channelId, content string) {
	// Convert all ` to ' to avoid code block formatting
	content = strings.ReplaceAll(content, "`", "'")

	// truncate the content to 2000 characters - the code block will add 6 characters plus the markdown label (2) and the newlines (2) = 10
	if len(content) > 2000-10 {
		content = content[:2000-10]
	}

	// Convert the content to a code block
	content = "```md\n" + content + "\n```"

	// Send a copy of the original content to the user
	ctx.Session().ChannelMessageSend(channelId, content)
}
