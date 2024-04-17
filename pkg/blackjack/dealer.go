package blackjack

import (
	"fmt"
	"sync"
	"time"
)

type GameStage string

const (
	IdleStage      GameStage = "Idle"
	JoinStage      GameStage = "Join"
	RoundStage     GameStage = "Playing"
	PayoutStage    GameStage = "Payout"
	ReshuffleStage GameStage = "Reshuffle"
	FinishedStage  GameStage = "Finished"
)

type StateChangeCallback func(stage GameStage, state GameState, messageId, channelId string)

type User struct {
	Id         string
	Hand       Hand
	InitialBet int64
	Bet        int64
	Blackjack  bool
}

type GameState struct {
	Id          string
	Shoe        Shoe
	Hand        Hand
	PlayerTurn  int
	Users       []User
	ShoePlayers map[string]bool
}

type Dealer struct {
	State GameState
	Stage GameStage

	action chan int

	messageId string
	channelId string

	// onStateChange is a callback function that is called when the game state
	// changes. This is useful for sending updates to the clients so they can
	// render the game state.
	onStateChange StateChangeCallback

	// onAchievement is a callback function that is called when a user unlocks an
	// achievement. This is useful for tracking user progress and notifying the
	// user of their achievement.
	onAchievement AchievementCallback

	mu sync.Mutex
}

func (d *Dealer) changeStage(stage GameStage) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.Stage = stage
	d.onStateChange(stage, d.State, d.messageId, d.channelId)
}

func (d *Dealer) commitState() {
	d.onStateChange(d.Stage, d.State, d.messageId, d.channelId)
}

func newState() GameState {
	s := GameState{
		Id:          fmt.Sprintf("%d", time.Now().UTC().Unix()),
		Shoe:        NewShoe(DefaultDeckCount),
		Hand:        make([]Card, 0),
		PlayerTurn:  -1,
		Users:       make([]User, 0),
		ShoePlayers: make(map[string]bool),
	}

	s.Shoe.Shuffle()
	return s
}
