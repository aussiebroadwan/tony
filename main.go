package main

import (
	"os"
	"os/signal"

	"github.com/aussiebroadwan/tony/commands"
	"github.com/aussiebroadwan/tony/database"
	"github.com/aussiebroadwan/tony/framework"
	"github.com/aussiebroadwan/tony/moderation"
	"github.com/aussiebroadwan/tony/pkg/reminders"

	"github.com/joho/godotenv"

	log "github.com/sirupsen/logrus"
)

const VERSION = "0.1.0"

var (
	token    = os.Getenv("DISCORD_TOKEN")
	serverId = os.Getenv("DISCORD_SERVER_ID")
)

func init() {
	// Setup logging
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	godotenv.Load()

	// Check if token is provided
	if token == "" {
		log.Fatal("No token provided. Please set DISCORD_TOKEN environment variable.")
		return
	}

	// Print version
	log.Infof("Tony v%s", VERSION)
}

func main() {
	// Setup database
	db := database.NewDatabase("tony.db")
	defer db.Close()

	// Create a new bot
	bot, err := framework.NewBot(token, serverId, db)
	if err != nil {
		log.Fatalf("Error creating bot: %s", err)
		return
	}

	// Setup reminders
	go reminders.Run()
	defer reminders.Stop()

	// Register routes
	bot.Register(
		framework.NewRoute(bot, "ping", &commands.PingCommand{}),

		framework.NewRoute(bot, "remind",
			&commands.RemindCommand{},

			// remind <subcommand>
			framework.NewSubRoute(bot, "add", &commands.RemindAddSubCommand{}),
			framework.NewSubRoute(bot, "del", &commands.RemindDeleteSubCommand{}),
			framework.NewSubRoute(bot, "list", &commands.RemindListSubCommand{}),
			framework.NewSubRoute(bot, "status", &commands.RemindStatusSubCommand{}),
		),
	)

	bot.DefineModerationRules(
		framework.Rule("tech-news", &moderation.ModerateNewsRule{}),
		framework.Rule("rss", &moderation.ModerateRSSRule{}),
	)

	// Run the bot
	if err = bot.Run(); err != nil {
		log.Fatalf("Error running bot: %s", err)
		return
	}
	defer bot.Close()

	// Setup reminders database, requires a discord session
	database.SetupRemindersDB(db, bot.Discord)

	waitForInterrupt()
}

func waitForInterrupt() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Shutting down...")
}
