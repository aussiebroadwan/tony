package main

import (
	"fmt"
	"os"
	"os/signal"

	app "github.com/aussiebroadwan/tony/applications"
	"github.com/aussiebroadwan/tony/applications/autopin"
	blackjack_app "github.com/aussiebroadwan/tony/applications/blackjack"
	"github.com/aussiebroadwan/tony/applications/remind"
	snailrace_app "github.com/aussiebroadwan/tony/applications/snailrace"
	walletApp "github.com/aussiebroadwan/tony/applications/wallet"
	"github.com/aussiebroadwan/tony/pkg/tradingcards"
	"github.com/aussiebroadwan/tony/pkg/wallet"
	"github.com/bwmarrin/discordgo"

	"github.com/aussiebroadwan/tony/database"
	"github.com/aussiebroadwan/tony/framework"

	log "github.com/sirupsen/logrus"
)

var (
	VERSION  = "Unreleased"
	SERVERID = ""
)

func init() {
	// Setup logging
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {
	if version := os.Getenv("TONY_VERSION"); version != "" {
		VERSION = version
	}

	// Print version
	log.Infof("Tony %s", VERSION)

	// Setup database
	db := database.NewDatabase()
	wallet.SetupWalletDB(db, log.WithField("src", "wallet"))
	tradingcards.SetupTradingCardsDB(db, log.WithField("src", "tradingcards"))

	token := os.Getenv("DISCORD_TOKEN")
	SERVERID = os.Getenv("DISCORD_SERVER_ID")

	// Check if token is provided
	if token == "" {
		log.Fatal("No token provided. Please set DISCORD_TOKEN environment variable.")
		return
	}

	if SERVERID == "" {
		log.Fatal("No server ID provided. Please set DISCORD_SERVER_ID environment variable.")
		return
	}

	// Create a new bot
	bot, err := framework.NewBot(token, SERVERID, db)
	if err != nil {
		log.Fatalf("Error creating bot: %s", err)
		return
	}

	bot.OnStartup(startupCb)

	// Register routes
	bot.Register(
		walletApp.RegisterWalletApp(bot),

		app.RegisterPingApp(bot),

		remind.RegisterRemindApp(bot),
		autopin.RegisterAutopinApp(bot),

		blackjack_app.RegisterBlackjackApp(bot),
		snailrace_app.RegisterSnailraceApp(bot),

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

func startupCb(ctx framework.StartupContext) {
	session := ctx.Session()

	// Get Channel ID
	channelName := os.Getenv("DISCORD_STARTUP_CHANNEL")
	if channelName == "" {
		channelName = "tony-dev"
	}

	// Get channels for this guild
	channels, _ := session.GuildChannels(SERVERID)
	for _, channel := range channels {
		if channel.Name == channelName {

			session.ChannelMessageSendEmbed(channel.ID, &discordgo.MessageEmbed{
				Title:       "Tony Status",
				Description: fmt.Sprintf("I'm now online running version `%s`. Checkout the changelog on my Github repo to see what's changed!\n\nhttps://github.com/aussiebroadwan/tony", VERSION),
				Color:       0x4CAF50,
			})

			return
		}
	}

	log.Warnf("Channel startup channel (#%s) not found", channelName)
}
