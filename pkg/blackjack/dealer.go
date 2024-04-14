package blackjack

import (
	"sync"
)

type GameStage int

const (
	IdleStage GameStage = iota
	JoinStage
	RoundStage
	PayoutStage
	ReshuffleStage
)

type User struct {
	Id        string
	Hand      Hand
	Bet       int64
	Blackjack bool
}

type GameState struct {
	Shoe       Shoe
	Hand       Hand
	PlayerTurn int
	Users      []User
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
	onStateChange func(stage GameStage, state GameState, messageId, channelId string)

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
		Shoe:       NewShoe(DefaultDeckCount),
		Hand:       make([]Card, 0),
		PlayerTurn: -1,
		Users:      make([]User, 0),
	}

	s.Shoe.Shuffle()
	return s
}
