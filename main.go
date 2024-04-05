package main

import (
	"os"
	"os/signal"

	rules "github.com/aussiebroadwan/tony/applicationRules"
	app "github.com/aussiebroadwan/tony/applications"

	"github.com/aussiebroadwan/tony/database"
	"github.com/aussiebroadwan/tony/framework"
	"github.com/aussiebroadwan/tony/pkg/reminders"

	log "github.com/sirupsen/logrus"
)

const VERSION = "0.1.0"

func init() {
	// Setup logging
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {
	// Print version
	log.Infof("Tony v%s", VERSION)

	// Setup database
	db := database.NewDatabase("tony.db")
	defer db.Close()

	token := os.Getenv("DISCORD_TOKEN")
	serverId := os.Getenv("DISCORD_SERVER_ID")

	// Check if token is provided
	if token == "" {
		log.Fatal("No token provided. Please set DISCORD_TOKEN environment variable.")
		return
	}

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
		framework.NewRoute(bot, "ping",
			// ping
			&app.PingCommand{},

			// ping button
			framework.NewSubRoute(bot, "button", &app.PingButtonCommand{}),
		),

		framework.NewRoute(bot, "remind",
			// remind
			&app.RemindCommand{}, // [NOP]

			// remind <subcommand>
			framework.NewSubRoute(bot, "add", &app.RemindAddSubCommand{}),
			framework.NewSubRoute(bot, "del", &app.RemindDeleteSubCommand{}),
			framework.NewSubRoute(bot, "list", &app.RemindListSubCommand{}),
			framework.NewSubRoute(bot, "status", &app.RemindStatusSubCommand{}),
		),
	)

	bot.RegisterRules(
		framework.Rule("tech-news", &rules.ModerateNewsRule{}),
		framework.Rule("rss", &rules.ModerateRSSRule{}),

		framework.Rule(".*", &rules.AutoPinRule{}),
	)

	// Run the bot
	if err = bot.Run(); err != nil {
		log.Fatalf("Error running bot: %s", err)
		return
	}
	defer bot.Close()

	// Setup reminders database, requires a discord session
	database.SetupRemindersDB(db, bot.Discord)
	database.SetupAutoPinDB(db)

	waitForInterrupt()
}

func waitForInterrupt() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Shutting down...")
}
