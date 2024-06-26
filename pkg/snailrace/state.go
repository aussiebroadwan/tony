package snailrace

import (
	"math/rand"
	"sort"
	"time"
)

// Constants to represent the states of a race.
const (
	StateJoining = iota
	StateBetting
	StateInProgress
	StateFinished
	StateCancelled

	MinRacers = 2
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

	Snails         []*Snail
	SnailPositions map[int][]float64
	Place          map[int]int
	RequriedSteps  int
	snailsToRemove []string

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
				if len(r.Snails) < MinRacers {
					r.transitionState(StateCancelled)
					return
				}

				r.puntersPlaceBets()
				r.transitionState(StateBetting)
			}
		case StateBetting:
			if time.Now().After(r.Race.StartAt) {
				r.SnailPositions = make(map[int][]float64)
				r.SimulateRace()
				r.transitionState(StateInProgress)
			}
		case StateInProgress:
			if r.Step < r.RequriedSteps {
				r.Step++
				r.stateCb(*r, r.MessageId, r.ChannelId)
			} else {
				r.updateSnailHistory()
				r.transitionState(StateFinished)
			}
		case StateFinished, StateCancelled:
			r.removeMarkedSnail()
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
func (r *RaceState) Join(snail Snail) error {
	snailPtr := &snail

	// check if the snail is already joined
	for _, s := range r.Snails {
		if s.Id == snailPtr.Id {
			return ErrAlreadyJoined
		}
	}

	r.Snails = append(r.Snails, snailPtr)
	r.Race.joinRace(snailPtr)
	return nil
}

func (r *RaceState) SimulateRace() {

	// Temporary struct to hold the results of the snails positions in the race
	type Result struct {
		SnailIndex int
		Positions  int
	}
	var results []Result
	r.RequriedSteps = 0

	// Build the race positions for the snails
	for i, snail := range r.Race.Snails {
		positions := snail.snail.SimulateRace(r.Race.Id)
		r.SnailPositions[i] = positions
		results = append(results, Result{SnailIndex: i, Positions: len(positions)})

		if len(positions) > r.RequriedSteps {
			r.RequriedSteps = len(positions)
		}
	}

	// Sort results by the number of positions, ascending (fewer positions means faster)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Positions < results[j].Positions
	})

	// Cache the final placements of the snails
	r.Place = make(map[int]int)
	for i, result := range results {
		r.Place[result.SnailIndex] = i + 1
	}
}

func (r *RaceState) updateSnailHistory() {
	for i := range r.Snails {
		var snail = Snail{Id: r.Snails[i].Id}
		database.First(&snail)

		// Update the snail's history
		snail.Prev3Place = snail.Prev2Place
		snail.Prev2Place = snail.Prev1Place
		snail.Prev1Place = r.Place[i]

		// Save the snail's updated history
		database.Save(&snail)
	}
}

func (r *RaceState) removeMarkedSnail() {
	snails := []Snail{}
	for _, id := range r.snailsToRemove {
		snails = append(snails, Snail{Id: id})
	}

	database.Unscoped().Delete(&snails)
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
