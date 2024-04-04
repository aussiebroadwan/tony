package framework

import (
	"context"
	"database/sql"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"

	_ "github.com/mattn/go-sqlite3"
)

type ContextKey string

const (
	CtxInteraction ContextKey = "interaction"
	CtxMessage     ContextKey = "message"
	CtxSession     ContextKey = "session"
	CtxDatabase    ContextKey = "db"
	CtxLogger      ContextKey = "logger"
	CtxEventValue  ContextKey = "event_val"
)

type ContextOpt func(*Context)

func withDatabase(db *sql.DB) ContextOpt {
	return func(c *Context) {
		c.ctx = context.WithValue(c.ctx, CtxDatabase, db)
	}
}

func withInteraction(i *discordgo.Interaction) ContextOpt {
	return func(c *Context) {
		c.ctx = context.WithValue(c.ctx, CtxInteraction, i)
	}
}

func withSession(s *discordgo.Session) ContextOpt {
	return func(c *Context) {
		c.ctx = context.WithValue(c.ctx, CtxSession, s)
	}
}

func withMessage(m *discordgo.Message) ContextOpt {
	return func(c *Context) {
		c.ctx = context.WithValue(c.ctx, CtxMessage, m)
	}
}

func withLogger(l *log.Entry) ContextOpt {
	return func(c *Context) {
		c.ctx = context.WithValue(c.ctx, CtxLogger, l)
	}
}

func withEventValue(v string) ContextOpt {
	return func(c *Context) {
		c.ctx = context.WithValue(c.ctx, CtxEventValue, v)
	}
}

// Context is a wrapper around the context.Context type that includes
// a reference to the discordgo.Message and discordgo.Session objects
// that triggered the command.
type Context struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
}

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
	return c.ctx.Value(CtxSession).(*discordgo.Session)
}

func (c *Context) Message() *discordgo.Message {
	return c.ctx.Value(CtxMessage).(*discordgo.Message)
}

func (c *Context) Interaction() *discordgo.Interaction {
	return c.ctx.Value(CtxInteraction).(*discordgo.Interaction)
}

func (c *Context) Database() *sql.DB {
	return c.ctx.Value(CtxDatabase).(*sql.DB)
}

func (c *Context) Logger() *log.Entry {
	return c.ctx.Value(CtxLogger).(*log.Entry)
}

func (c *Context) EventValue() string {
	return c.ctx.Value(CtxLogger).(string)
}
