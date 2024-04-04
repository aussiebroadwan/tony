package framework

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"

	log "github.com/sirupsen/logrus"

	_ "github.com/mattn/go-sqlite3"
)

type Bot struct {
	Discord *discordgo.Session

	serverId string
	Routes   []Route

	lg *log.Entry
	db *sql.DB
}

func NewBot(token string, serverId string, db *sql.DB) (*Bot, error) {
	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		Discord: discord,

		serverId: serverId,
		Routes:   make([]Route, 0),
		lg:       log.WithField("src", "bot"),
		db:       db,
	}, nil
}

// Register adds routes to the bot
func (b *Bot) Register(routes ...Route) {
	b.Routes = append(b.Routes, routes...)
}

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

func (b *Bot) registerAllCommandsAndRouting() {
	// Register the route with Discord
	for _, route := range b.Routes {
		createdApp, err := b.Discord.ApplicationCommandCreate(b.Discord.State.User.ID, b.serverId, route.declaration)
		if err != nil {
			b.lg.Errorf("Error creating command: %s", err)
			continue
		}
		route.declaration = createdApp

		for k := range route.commandRoute {
			b.lg.Infof("Registered command route: %s", k)
		}
	}

	// Define a function to build the route key
	appKeyBuilder := func(i *discordgo.InteractionCreate) string {
		routeKey := i.ApplicationCommandData().Name

		// Base case: If there are no options of type SubCommand, return the route key
		if len(i.ApplicationCommandData().Options) == 0 {
			return routeKey
		}

		// Recursive case: If there are options of type SubCommand, append the option name to the route key
		if i.ApplicationCommandData().Options[0].Type != discordgo.ApplicationCommandOptionSubCommand {
			return routeKey
		}

		return fmt.Sprintf("%s.%s", routeKey, keyBuilder(i.ApplicationCommandData().Options[0]))
	}

	// Handle the route execution
	b.Discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		var routeKey string = ""
		var eventValue string = ""

		// Route Key is the name of the command dot separated by the subcommands
		// e.g. "remind" or "remind.add". It is slightly different for the
		// message components and modals as they should have the CustomID with
		// the format of "command.subcommand:value" so the route key is the
		// part before the colon.
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			routeKey = appKeyBuilder(i)
		case discordgo.InteractionMessageComponent:
			routeKey, eventValue, _ = strings.Cut(i.MessageComponentData().CustomID, ":")
		case discordgo.InteractionModalSubmit:
			routeKey, eventValue, _ = strings.Cut(i.ModalSubmitData().CustomID, ":")
		default:
			b.lg.Errorf("Unknown interaction type: %s", i.Type.String())
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Unknwon Interaction",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		// Create a new context for the route
		ctx := NewContext(
			withSession(s),
			withDatabase(b.db),
			withInteraction(i.Interaction),
			withMessage(i.Interaction.Message),
			withLogger(b.lg.WithFields(log.Fields{
				"route": routeKey,
				"type":  i.Type.String(),
			})),
		)

		// Find the route
		for _, route := range b.Routes {
			if er, ok := route.commandRoute[routeKey]; ok {

				// If the route is found and it is just a command, execute it
				if i.Type == discordgo.InteractionApplicationCommand && (er.GetType() == CommandTypeApp || er.GetType() == CommandTypeAppAndEvent) {
					b.lg.Infof("Executing command: %s", routeKey)
					er.Execute(ctx)
					return
				}

				// If the route is found and it is an event handler, execute it
				if i.Type != discordgo.InteractionApplicationCommand && (er.GetType() == CommandTypeEvent || er.GetType() == CommandTypeAppAndEvent) {
					// Set the event value for the route
					withEventValue(eventValue)(ctx)

					// If the route is found and it is an event handler, execute it
					b.lg.Infof("Executing event: %s", routeKey)
					er.OnEvent(ctx, i.Type)
					return
				}
			}
		}

		// If the route is not found, respond with an error message
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Interaction not found",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	})
}

func (b *Bot) deregisterAllCommands() {

	// Delete the route from Discord
	for _, route := range b.Routes {
		b.Discord.ApplicationCommandDelete(b.Discord.State.User.ID, b.serverId, route.declaration.ID)

		for k := range route.commandRoute {
			b.lg.Infof("Deregistered command route: %s", k)
		}
	}
}

func (b *Bot) DefineModerationRules(rules ...ActionableRule) {
	b.Discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		// Get Channel Name from Message
		channel, err := s.State.Channel(m.ChannelID)
		if err != nil {
			return
		}

		// Test a regex match for the channel name against the rule
		for _, rule := range rules {
			if match, _ := regexp.Match(rule.Channel, []byte(channel.Name)); match {

				// Test the rule
				if err := rule.Rule.Test(m.Content); err != nil {
					rule.Rule.Action(NewContext(
						withSession(s),
						withInteraction(nil), // No interaction for messages
						withMessage(m.Message),
						withLogger(b.lg.WithField("rule", rule.Rule.Name())),
						withDatabase(b.db),
					), err)
				}
			}
		}
	})
}

func (b *Bot) Run() error {
	if err := b.Discord.Open(); err != nil {
		return err
	}
	b.registerAllCommandsAndRouting()
	return nil
}

func (b *Bot) Close() error {
	b.deregisterAllCommands()
	return b.Discord.Close()
}
