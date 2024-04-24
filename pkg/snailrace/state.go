package snailrace

import (
	"math/rand"
	"time"
)

// Constants to represent the states of a race.
const (
	StateJoining = iota
	StateBetting
	StateInProgress
	StateFinished
	StateCancelled
)

// AchievementCallback defines a callback function type for when an achievement
// is unlocked. Important as we dont want this package to depend on any other
// package.
type AchievementCallback func(userId string, achievementName string) bool

// StateChangeCallback defines a function type that handles changes in race
// state.
type StateChangeCallback func(raceState RaceState, messageId, channelId string)

// RaceState holds all data related to the state of a race.
type RaceState struct {
	Race  *Race
	State int
	Step  int

	Snails []*Snail

	MessageId string
	ChannelId string

	stateCb       StateChangeCallback
	achievementCb AchievementCallback
}

func (r *RaceState) Start(betTime time.Time) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	r.stateCb(*r, r.MessageId, r.ChannelId)

	for range ticker.C {
		switch r.State {
		case StateJoining:
			if time.Now().After(betTime) {
				if len(r.Snails) == 0 {
					r.transitionState(StateCancelled)
					return
				}

				r.transitionState(StateBetting)
				r.puntersPlaceBets()
			}
		case StateBetting:
			if time.Now().After(r.Race.StartAt) {
				r.transitionState(StateInProgress)
			}
		case StateInProgress:
			if r.Step < 100 {
				r.Step++
				r.stateCb(*r, r.MessageId, r.ChannelId)
			} else {
				r.transitionState(StateFinished)
			}
		case StateFinished, StateCancelled:
			return
		}
	}
}

// transitionState updates the state of the race and triggers a callback.
func (r *RaceState) transitionState(newState int) {
	r.State = newState
	r.stateCb(*r, r.MessageId, r.ChannelId)
}

// Join adds a snail to the race.
func (r *RaceState) Join(snail Snail) {
	snailPtr := &snail
	r.Snails = append(r.Snails, snailPtr)
	r.Race.joinRace(snailPtr)
}

// puntersPlaceBets handles the betting process for all punters.
func (r *RaceState) puntersPlaceBets() {
	var punters []Punter
	database.Find(&punters) // Assuming database.Find abstracts error handling.

	shuffledPunters := shufflePunters(punters)
	if len(shuffledPunters) > PuntersPerRace {
		shuffledPunters = shuffledPunters[:PuntersPerRace]
	}

	r.Race.puntersPlaceBets(shuffledPunters)
}

// shufflePunters randomly shuffles a slice of punters.
func shufflePunters(punters []Punter) []Punter {
	rand.Shuffle(len(punters), func(i, j int) {
		punters[i], punters[j] = punters[j], punters[i]
	})
	return punters
}
