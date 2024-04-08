package main

import (
	"os"
	"os/signal"

	app "github.com/aussiebroadwan/tony/applications"
	"github.com/aussiebroadwan/tony/applications/autopin"
	"github.com/aussiebroadwan/tony/applications/remind"

	"github.com/aussiebroadwan/tony/database"
	"github.com/aussiebroadwan/tony/framework"

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

	// Register routes
	bot.Register(
		app.RegisterPingApp(bot),

		remind.RegisterRemindApp(bot),
		autopin.RegisterAutopinApp(bot),

		app.RegisterNewsModeration(bot),
		app.RegisterRSSModeration(bot),
	)

	// Run the bot
	if err = bot.Run(); err != nil {
		log.Fatalf("Error running bot: %s", err)
		return
	}
	defer bot.Close()

	waitForInterrupt()
}

func waitForInterrupt() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Shutting down...")
}
