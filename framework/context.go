package framework

import (
	"context"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ContextKey string

const (
	ctxInteraction ContextKey = "interaction"
	ctxMessage     ContextKey = "message"
	ctxSession     ContextKey = "session"
	ctxDatabase    ContextKey = "db"
	ctxLogger      ContextKey = "logger"
	ctxEventValue  ContextKey = "event_val"

	ctxReactionValue ContextKey = "reaction_val"
	ctxReactionAdd   ContextKey = "reaction_add"
)

type MountContext interface {
	Session() *discordgo.Session
	Database() *gorm.DB
	Logger() *log.Entry
}

type CommandContext interface {
	Session() *discordgo.Session
	Message() *discordgo.Message
	Interaction() *discordgo.Interaction
	GetOption(string) *discordgo.ApplicationCommandInteractionDataOption
	GetUser() *discordgo.User
	Database() *gorm.DB
	Logger() *log.Entry
}

type EventContext interface {
	Session() *discordgo.Session
	Message() *discordgo.Message
	Interaction() *discordgo.Interaction
	Database() *gorm.DB
	Logger() *log.Entry
	EventValue() string
}

type MessageContext interface {
	Session() *discordgo.Session
	Message() *discordgo.Message
	Database() *gorm.DB
	Logger() *log.Entry
}

type ReactionContext interface {
	Session() *discordgo.Session
	Database() *gorm.DB
	Logger() *log.Entry
	Reaction() (*discordgo.MessageReaction, bool)
}

// Context is a wrapper around the context.Context type that includes
// a reference to the discordgo.Message and discordgo.Session objects
// that triggered the command.
type Context struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
}
type ContextOpt func(*Context)

func NewContext(opts ...ContextOpt) *Context {

	// Create a new context with a cancel function
	ctx, cancel := context.WithCancel(context.Background())
	dContext := &Context{
		ctx:        ctx,
		cancelFunc: cancel,
	}

	// Apply the provided options to the context
	for _, opt := range opts {
		opt(dContext)
	}

	return dContext
}

func (c *Context) Session() *discordgo.Session {
	return c.ctx.Value(ctxSession).(*discordgo.Session)
}

func (c *Context) Message() *discordgo.Message {
	return c.ctx.Value(ctxMessage).(*discordgo.Message)
}

func (c *Context) Interaction() *discordgo.Interaction {
	return c.ctx.Value(ctxInteraction).(*discordgo.Interaction)
}

func (c *Context) GetUser() *discordgo.User {
	interaction := c.Interaction()
	user := interaction.User
	if user == nil {
		user = interaction.Member.User
	}
	return user
}

func (c *Context) GetOption(name string) *discordgo.ApplicationCommandInteractionDataOption {
	interaction := c.Interaction()
	options := interaction.ApplicationCommandData().Options[0].Options

	for _, opt := range options {
		if opt.Name == name {
			return opt
		}
	}
	return nil
}

func (c *Context) Database() *gorm.DB {
	return c.ctx.Value(ctxDatabase).(*gorm.DB)
}

func (c *Context) Logger() *log.Entry {
	return c.ctx.Value(ctxLogger).(*log.Entry)
}

func (c *Context) EventValue() string {
	return c.ctx.Value(ctxEventValue).(string)
}

func (c *Context) Reaction() (*discordgo.MessageReaction, bool) {
	val, add := c.ctx.Value(ctxReactionValue), c.ctx.Value(ctxReactionAdd)
	return val.(*discordgo.MessageReaction), add.(bool)
}

func withDatabase(db *gorm.DB) ContextOpt {
	return func(c *Context) {
		c.ctx = context.WithValue(c.ctx, ctxDatabase, db)
	}
}

func withInteraction(i *discordgo.Interaction) ContextOpt {
	return func(c *Context) {
		c.ctx = context.WithValue(c.ctx, ctxInteraction, i)
	}
}

func withSession(s *discordgo.Session) ContextOpt {
	return func(c *Context) {
		c.ctx = context.WithValue(c.ctx, ctxSession, s)
	}
}

func withMessage(m *discordgo.Message) ContextOpt {
	return func(c *Context) {
		c.ctx = context.WithValue(c.ctx, ctxMessage, m)
	}
}

func withLogger(l *log.Entry) ContextOpt {
	return func(c *Context) {
		c.ctx = context.WithValue(c.ctx, ctxLogger, l)
	}
}

func withEventValue(v string) ContextOpt {
	return func(c *Context) {
		c.ctx = context.WithValue(c.ctx, ctxEventValue, v)
	}
}

func withReaction(r *discordgo.MessageReaction, add bool) ContextOpt {
	return func(c *Context) {
		c.ctx = context.WithValue(c.ctx, ctxReactionValue, r)
		c.ctx = context.WithValue(c.ctx, ctxReactionAdd, add)
	}
}
