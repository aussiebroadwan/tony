package blackjack

import (
	"fmt"
	"sync"
	"time"
)

const (
	MaxPlayers            = 7
	DefaultDeckCount      = 6
	DefaultPayoutRatio    = 1.0
	BlackjackPayoutRatio  = 1.5
	JoinTimeoutDuration   = 30 * time.Second
	CardDealInterval      = 500 * time.Millisecond
	PlayerTurnTimeout     = 15 * time.Second
	MaximumHandScore      = 21
	DealerStandScore      = 17
	PayoutProcessingDelay = 5 * time.Second
)

// The dealer manages the game state and controls the flow of the game. This is
// the only dealer instance that should be used in the application.
var dealer *Dealer = &Dealer{
	State: GameState{
		Shoe:       NewShoe(DefaultDeckCount),
		Hand:       make([]Card, 0),
		Users:      make([]User, 0),
		PlayerTurn: 0,
	},
	Stage:         IdleStage,
	onStateChange: func(stage GameStage, state GameState, messageId, channelId string) {},
	action:        make(chan int),
	mu:            sync.Mutex{},
}

// initialDeal deals two cards to each player and one to the dealer.
func initialDeal() {
	dealer.mu.Lock()
	defer dealer.mu.Unlock()

	// Deal first card for the dealer.
	dealer.State.Hand = append(dealer.State.Hand, dealer.State.Shoe.Draw())
	dealer.commitState()
	time.Sleep(CardDealInterval)

	// Deal two cards to each player.
	for i := 0; i < 2; i++ {
		for index := range dealer.State.Users {
			user := &dealer.State.Users[index]
			user.Hand = append(user.Hand, dealer.State.Shoe.Draw())
			checkForBlackjack(user)

			dealer.commitState()
			time.Sleep(CardDealInterval)
		}
	}
}

// checkForBlackjack checks if the user has a blackjack.
func checkForBlackjack(user *User) {
	if user.Hand.Score() == MaximumHandScore {
		user.Blackjack = true
	}
}

// calculatePayouts determines the winnings or losses for each player.
func calculatePayouts() {
	dealer.mu.Lock()
	defer dealer.mu.Unlock()

	for index := range dealer.State.Users {
		user := &dealer.State.Users[index]
		if user.Hand.Score() > MaximumHandScore {
			user.Bet = 0 // Player busts
			continue
		}

		if user.Blackjack {
			user.Bet += int64(float64(user.Bet) * BlackjackPayoutRatio)
		} else if user.Hand.Score() > dealer.State.Hand.Score() {
			user.Bet += int64(float64(user.Bet) * DefaultPayoutRatio)
		} else if user.Hand.Score() == dealer.State.Hand.Score() {
			// Push: no change to bet
		} else {
			user.Bet = 0 // Player loses
		}
	}
}

// processPlayerTurns cycles through each player's turn until all have acted.
func processPlayerTurns() {
	dealer.State.PlayerTurn = 0
	for dealer.State.PlayerTurn < len(dealer.State.Users) {
		ticker := time.NewTicker(PlayerTurnTimeout)

		select {
		case <-dealer.action:
			ticker.Stop() // Player has taken an action, stop the timer.

		case <-ticker.C:
			// Time expired, assume stand if no action taken.
			dealer.mu.Lock()
			dealer.State.PlayerTurn++
			dealer.mu.Unlock()

			ticker.Stop()
		}
	}

	// Process the dealer's turn
	dealerPlay()
}

// dealerPlay simulates the dealer's play according to the house rules.
func dealerPlay() {
	for dealer.State.Hand.Score() < DealerStandScore {
		dealer.State.Hand = append(dealer.State.Hand, dealer.State.Shoe.Draw())
		dealer.commitState()
		time.Sleep(CardDealInterval)
	}
}

// executeGameLoop manages the flow of the game from start to finish.
func executeGameLoop() {
	fmt.Println("Starting game loop")
	dealer.changeStage(JoinStage)
	defer dealer.changeStage(IdleStage)
	fmt.Println("Ready to join")

	time.Sleep(JoinTimeoutDuration)
	fmt.Println("Join timeout expired")
	if len(dealer.State.Users) < 1 {
		fmt.Println("No Users")
		return // Not enough players to start the game.
	}

	initialDeal()
	fmt.Println("Initial Deal")
	dealer.changeStage(RoundStage)
	processPlayerTurns()

	calculatePayouts()
	dealer.changeStage(PayoutStage)
	time.Sleep(PayoutProcessingDelay)

	// Prepare for the next round or end the game if no players.
	executeGameLoop()
}
