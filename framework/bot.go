package framework

import (
	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"
)

type Bot struct {
	Discord *discordgo.Session

	serverId string
	Routes   []Route

	registeredDiscordApplications map[string]*discordgo.ApplicationCommand

	lg *log.Entry
	db *gorm.DB
}

func NewBot(token string, serverId string, db *gorm.DB) (*Bot, error) {
	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		Discord: discord,

		serverId: serverId,
		Routes:   make([]Route, 0),

		registeredDiscordApplications: make(map[string]*discordgo.ApplicationCommand),

		lg: log.WithField("src", "bot"),
		db: db,
	}, nil
}

// Register adds routes to the bot
func (b *Bot) Register(routes ...Route) {
	b.Routes = append(b.Routes, routes...)

	for _, route := range routes {
		// Register the route with the bot
		for key := range route.appRoute {
			b.lg.WithField("app", route.Name).Infof("Registering route: %s", key)
		}
	}
}

func (b *Bot) registerDiscordApplicationCommands() {

	// Register the route with Discord
	for _, route := range b.Routes {

		// Register Application if it is a command
		if route.App.GetType()&AppTypeCommand != 0 {
			app := route.App.(ApplicationCommand)

			// Register the command with Discord
			createdApp, err := b.Discord.ApplicationCommandCreate(
				b.Discord.State.User.ID,
				b.serverId,
				app.GetDefinition(),
			)

			// Check for errors
			if err != nil {
				b.lg.WithField("app", app.GetDefinition().Name).Errorf("Error creating command: %s", err)
				continue
			}

			// Add the application to the registered applications list for
			// deregistration later
			b.registeredDiscordApplications[route.Name] = createdApp

			// Notify that the application has been registered
			b.lg.Infof("Created Discord Application: %s", route.Name)
		}
	}
}

func (b *Bot) interactionCreateHandler() func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Fetch the route key and event value from the interaction
		routeKey, eventValue, err := GetRouteKey(i)
		if err != nil {
			b.lg.WithError(err).Errorf("Unknown interaction type")
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Unknwon Interaction",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		// Get the user from the interaction
		user := i.Interaction.Member.User
		if user == nil {
			user = i.Interaction.User
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
				"user":  user.ID,
			})),
		)

		// Find the route
		for _, route := range b.Routes {
			if er, ok := route.appRoute[routeKey]; ok {

				// If the route is found and it is just a command, execute it
				if i.Type == discordgo.InteractionApplicationCommand && (er.GetType()&(AppTypeCommand) != 0) {
					b.lg.Infof("Executing command: %s", routeKey)
					er.(ApplicationCommand).OnCommand(ctx)
					return
				}

				if i.Type == discordgo.InteractionApplicationCommand && (er.GetType()&(AppTypeSubCommand) != 0) {
					b.lg.Infof("Executing subcommand: %s", routeKey)
					er.(ApplicationSubCommand).OnCommand(ctx)
					return
				}

				// If the route is found and it is an event handler, execute it
				if i.Type != discordgo.InteractionApplicationCommand && (er.GetType()&AppTypeEvent != 0) {
					// Set the event value for the route
					withEventValue(eventValue)(ctx)

					b.lg.Infof("Executing event: %s", routeKey)
					er.(ApplicationEvent).OnEvent(ctx, i.Type)
					return
				}
			}
		}

		ctx.Logger().Errorf("Interaction not found")

		// If the route is not found, respond with an error message
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Interaction not found",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
}

func (b *Bot) messageCreateHandler() func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		// Get Channel Name from Message
		channel, err := s.State.Channel(m.ChannelID)
		if err != nil {
			return
		}

		// Get the user from the message creation
		user := m.Message.Author
		if user == nil {
			user = m.Message.Member.User
		}

		// Create a new context for the route
		ctx := NewContext(
			withSession(s),
			withMessage(m.Message),
			withDatabase(b.db),
		)

		// Test a regex match for the channel name against the rule
		for _, route := range b.Routes {

			// Check if the route is a message route
			if route.App.GetType()&AppTypeMessage == 0 {
				continue
			}

			// Set the logger for the app
			withLogger(b.lg.WithFields(log.Fields{
				"route": route.Name,
				"type":  "message",
				"user":  user.ID,
			}))(ctx)

			// Execute the app
			route.App.(ApplicationMessage).OnMessage(ctx, channel)
		}
	}
}

func (b *Bot) reactionAddHandler() func(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	return func(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
		if r.UserID == s.State.User.ID {
			return
		}

		ctx := NewContext(
			withSession(s),
			withDatabase(b.db),
			withReaction(r.MessageReaction, true),
		)

		// Get the user from the reaction
		user := r.Member.User

		// Test a regex match for the channel name against the rule
		for _, route := range b.Routes {

			// Check if the route is a reaction route
			if route.App.GetType()&AppTypeReaction == 0 {
				continue
			}

			// Create a new context for the route
			withLogger(b.lg.WithFields(log.Fields{
				"route": route.Name,
				"type":  "reaction_add",
				"user":  user.ID,
			}))(ctx)

			// Execute the app
			route.App.(ApplicationReaction).OnReaction(ctx)
		}
	}
}

func (b *Bot) reactionRemoveHandler() func(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	return func(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
		if r.UserID == s.State.User.ID {
			return
		}

		ctx := NewContext(
			withSession(s),
			withDatabase(b.db),
			withReaction(r.MessageReaction, false),
		)

		// Test a regex match for the channel name against the rule
		for _, route := range b.Routes {

			// Check if the route is a reaction route
			if route.App.GetType()&AppTypeReaction == 0 {
				continue
			}

			// Create a new context for the route
			withLogger(b.lg.WithFields(log.Fields{
				"route": route.Name,
				"type":  "reaction_remove",
				"user":  r.UserID,
			}))(ctx)

			// Execute the app
			route.App.(ApplicationReaction).OnReaction(ctx)
		}
	}
}

func (b *Bot) registerAllCommandsAndRouting() {
	b.registerDiscordApplicationCommands()

	// Handle the route execution
	b.Discord.AddHandler(b.interactionCreateHandler())
	b.Discord.AddHandler(b.messageCreateHandler())
	b.Discord.AddHandler(b.reactionAddHandler())
	b.Discord.AddHandler(b.reactionRemoveHandler())
}

func (b *Bot) deregisterAllCommands() {
	// Delete the route from Discord
	for route, app := range b.registeredDiscordApplications {
		b.Discord.ApplicationCommandDelete(b.Discord.State.User.ID, b.serverId, app.ID)

		b.lg.Infof("Deregistered Application: %s", route)
	}

	// Clear the registered applications
	b.registeredDiscordApplications = make(map[string]*discordgo.ApplicationCommand)
}

func (b *Bot) Run() error {
	if err := b.Discord.Open(); err != nil {
		return err
	}
	b.registerAllCommandsAndRouting()

	ctx := NewContext(
		withSession(b.Discord),
		withDatabase(b.db),
	)

	// Run the OnMount function for each route
	for _, route := range b.Routes {

		withLogger(b.lg.WithFields(log.Fields{
			"route": route.Name,
			"type":  "mount",
		}))(ctx)

		if route.App.GetType()&AppTypeMountable != 0 {
			route.App.(ApplicationMountable).OnMount(ctx)
		}

		for _, subroute := range route.Subroutes {
			withLogger(b.lg.WithFields(log.Fields{
				"route": subroute.Name,
				"type":  "mount",
			}))(ctx)

			if subroute.App.GetType()&AppTypeMountable != 0 {
				subroute.App.(ApplicationMountable).OnMount(ctx)
			}
		}
	}

	return nil
}

func (b *Bot) Close() error {
	b.deregisterAllCommands()
	return b.Discord.Close()
}
