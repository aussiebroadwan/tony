package framework

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func keyBuilder(opt *discordgo.ApplicationCommandInteractionDataOption) string {
	routeKey := opt.Name

	// Base case: If there are no options of type SubCommand, return the route key
	if len(opt.Options) == 0 {
		return routeKey
	}
	if opt.Options[0].Type != discordgo.ApplicationCommandOptionSubCommand {
		return routeKey
	}

	// Recursive case: If there are options of type SubCommand, append the option name to the route key
	return fmt.Sprintf("%s.%s", routeKey, keyBuilder(opt.Options[0]))
}

func routeBuilder(i *discordgo.InteractionCreate) string {
	routeKey := i.ApplicationCommandData().Name

	// Base case: If there are no options of type SubCommand, return the route key
	if len(i.ApplicationCommandData().Options) == 0 {
		return routeKey
	}

	// Recursive case: If there are options of type SubCommand, append the option name to the route key
	if i.ApplicationCommandData().Options[0].Type != discordgo.ApplicationCommandOptionSubCommand && i.ApplicationCommandData().Options[0].Type != discordgo.ApplicationCommandOptionSubCommandGroup {
		return routeKey
	}

	return fmt.Sprintf("%s.%s", routeKey, keyBuilder(i.ApplicationCommandData().Options[0]))
}

// Route Key is the name of the command dot separated by the subcommands
// e.g. "remind" or "remind.add". It is slightly different for the
// message components and modals as they should have the CustomID with
// the format of "command.subcommand:value" so the route key is the
// part before the colon.
func GetRouteKey(i *discordgo.InteractionCreate) (routeKey, eventValue string, err error) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		routeKey = routeBuilder(i)
	case discordgo.InteractionMessageComponent:
		routeKey, eventValue, _ = strings.Cut(i.MessageComponentData().CustomID, ":")
	case discordgo.InteractionModalSubmit:
		routeKey, eventValue, _ = strings.Cut(i.ModalSubmitData().CustomID, ":")
	default:
		return "", "", fmt.Errorf("interaction type %s not supported", i.Type.String())
	}

	return
}
