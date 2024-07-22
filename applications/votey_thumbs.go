package applications

import (
	"github.com/aussiebroadwan/tony/framework"
	"github.com/bwmarrin/discordgo"
)

func RegisterVoteyThumbsApp(bot *framework.Bot) framework.Route {
	return framework.NewRoute(bot, "voteythumbs",
		// voteythumbs
		&VoteyThumbsCommand{},
	)
}

type VoteyThumbsCommand struct {
	framework.ApplicationCommand
}

func (vtc VoteyThumbsCommand) GetType() framework.AppType {
	return framework.AppTypeCommand
}

func (vtc VoteyThumbsCommand) GetDefinition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "voteythumbs",
		Description: "Respectfully debate things that probably don't matter",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "question",
				Description: "The topic of debate",
				Required:    true,
			},
		},
	}
}

func (vtc VoteyThumbsCommand) OnCommand(ctx framework.CommandContext) {
	interaction := ctx.Interaction()
	commandOptions := interaction.ApplicationCommandData().Options

	// Get question that the user gave
	question, err := framework.GetOption(commandOptions, "question")
	if err != nil {
		ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "**Error:** question is required",
			},
		})
		return
	}

	// Send the question
	ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: question.StringValue(),
		},
	})

	// Now get the ID for the interaction and react to it.
	response, err := ctx.Session().InteractionResponse(ctx.Interaction())
	if err != nil {
		// If this happens, real chaos
		ctx.Logger().Warn("Failed to get interaction response on VoteyThumbs.")
		return
	}

	ctx.Session().MessageReactionAdd(response.ChannelID, response.ID, "👍")
	ctx.Session().MessageReactionAdd(response.ChannelID, response.ID, "👎")
}
