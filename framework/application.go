package framework

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type AppType int

const (
	// AppTypeNOP is a command that does nothing and is used as a
	// routing placeholder for parent containers for their sub applications.
	AppTypeNOP AppType = 1

	// AppTypeCommand is an discord application command that is executed by a
	// user and handled by the OnCommand() handler
	AppTypeCommand AppType = 1 << 1

	// AppTypeCommand is an discord application command that is executed by a
	// user and handled by the OnCommand() handler
	AppTypeSubCommand AppType = 1 << 2

	// AppTypeEvent is an event based application which handles user
	// interactions with the OnEvent() handler
	// message components or modals
	AppTypeEvent AppType = 1 << 3

	// AppTypeMessage is an application which runs on messages, handled by the
	// OnMessage() handler
	AppTypeMessage AppType = 1 << 4

	// AppTypeReaction is an application which runs on reactions, handled by the
	// OnReaction() handler
	AppTypeReaction AppType = 1 << 5

	// AppTypeMountable is an application which runs on mount, handled by the
	// OnMount() handler
	AppTypeMountable AppType = 1 << 6
)

type Application interface {
	GetType() AppType
}

type ApplicationMountable interface {
	Application
	OnMount(ctx MountContext)
}

type ApplicationCommand interface {
	Application
	GetDefinition() *discordgo.ApplicationCommand
	OnCommand(ctx CommandContext)
}

type ApplicationSubCommand interface {
	Application
	OnCommand(ctx CommandContext)
}

type ApplicationEvent interface {
	Application
	OnEvent(ctx EventContext, eventType discordgo.InteractionType)
}

type ApplicationMessage interface {
	Application
	OnMessage(ctx MessageContext, channel *discordgo.Channel)
}

type ApplicationReaction interface {
	Application
	OnReaction(ctx ReactionContext)
}

// Route associates a command name with a command instance and optional subcommands
type Route struct {
	Name      string
	App       Application
	Subroutes []Route

	appRoute map[string]Application
}

// NewRoute constructs a new Route
func NewRoute(bot *Bot, routeName string, command Application, subroutes ...Route) Route {
	r := Route{
		Name:      routeName,
		App:       command,
		Subroutes: subroutes,
		appRoute:  make(map[string]Application),
	}

	// Add the subroutes to the route map
	for _, sr := range subroutes {
		for k, v := range sr.appRoute {
			r.appRoute[fmt.Sprintf("%s.%s", r.Name, k)] = v
		}
	}

	// Get the Application Type and Check if the Type is implemented
	if !checkApplicationImplementation(bot, command, routeName) {
		bot.lg.Fatalf("Command %s is not properly implemented", routeName)
	}

	// Add the command to the command route
	r.appRoute[routeName] = command

	return r
}

func checkApplicationImplementation(bot *Bot, app Application, name string) bool {
	implements := true

	// Check if the app implements the Application interfaces
	_, implementesAppCommand := app.(ApplicationCommand)
	_, implementesAppSubCommand := app.(ApplicationSubCommand)
	_, implementesAppEvent := app.(ApplicationEvent)
	_, implementesAppMessage := app.(ApplicationMessage)
	_, implementesAppReaction := app.(ApplicationReaction)

	// Check if the app says its an Application Command but does not implement
	// the ApplicationCommand interface
	if app.GetType()&AppTypeCommand != 0 && !implementesAppCommand {
		bot.lg.Errorf("Command %s does not implement ApplicationCommand interface", name)
		implements = false
	}

	// Check if the app says its an Application SubCommand but does not implement
	// the ApplicationSubCommand interface
	if app.GetType()&AppTypeSubCommand != 0 && !implementesAppSubCommand {
		bot.lg.Errorf("SubCommand %s does not implement ApplicationSubCommand interface", name)
		implements = false
	}

	// Check if the app says its an Application Event but does not implement
	// the ApplicationEvent interface
	if app.GetType()&AppTypeEvent != 0 && !implementesAppEvent {
		bot.lg.Errorf("Event %s does not implement ApplicationEvent interface", name)
		implements = false
	}

	// Check if the app says its an Application Message but does not implement
	// the ApplicationMessage interface
	if app.GetType()&AppTypeMessage != 0 && !implementesAppMessage {
		bot.lg.Errorf("Message %s does not implement ApplicationMessage interface", name)
		implements = false
	}

	// Check if the app says its an Application Reaction but does not implement
	// the ApplicationReaction interface
	if app.GetType()&AppTypeReaction != 0 && !implementesAppReaction {
		bot.lg.Errorf("Reaction %s does not implement ApplicationReaction interface", name)
		implements = false
	}

	return implements
}

func GetOption(opts []*discordgo.ApplicationCommandInteractionDataOption, key string) (*discordgo.ApplicationCommandInteractionDataOption, error) {
	for _, opt := range opts {
		if opt.Name == key {
			return opt, nil
		}
	}
	return nil, fmt.Errorf("option %s not found", key)
}
