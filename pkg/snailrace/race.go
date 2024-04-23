package snailrace

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Race struct represents a single racing event.
type Race struct {
	Id string `gorm:"primaryKey"`

	StartAt time.Time // Time when the race is scheduled to start
	Pool    int64     // Total amount of money in the pool from the Punters

	mu       sync.Mutex      `gorm:"-"`
	Snails   []SnailRaceLink `gorm:"-"`
	UserBets []UserBet       `gorm:"-"`
}

// UserBet struct represents a bet made by a user on a snail.
type UserBet struct {
	UserId     string
	Amount     int64
	SnailIndex int
}

// SnailRaceLink struct links a snail to a race with a specific betting pool.
type SnailRaceLink struct {
	gorm.Model

	RaceId  string
	SnailId string
	snail   *Snail `gorm:"-"`

	Pool int64 // Total amount of money in the pool for this Snail from the Punters
}

// calculateOdds computes the betting odds for a snail based on the total pool
// and snail's pool.
func calculateOdds(racePool, snailPool int64) float64 {
	return float64(snailPool) / float64(racePool)
}

// newRace creates and initialises a new race with a unique identifier.
func newRace() *Race {
	id := uuid.New().String()[24:] // Truncating UUID to get a shorter ID.

	r := &Race{
		Id:       id,
		Pool:     0,
		Snails:   make([]SnailRaceLink, 0),
		UserBets: make([]UserBet, 0),
		mu:       sync.Mutex{},
	}
	database.Create(r)
	return r
}

// joinRace adds a snail to the race and registers it in the database.
func (r *Race) joinRace(s *Snail) {
	r.mu.Lock()
	defer r.mu.Unlock()

	link := SnailRaceLink{
		RaceId:  r.Id,
		SnailId: s.Id,
		snail:   s,
		Pool:    0,
	}
	database.Create(&link) // Register the new link in the database.
	r.Snails = append(r.Snails, link)
}

// placeBet records a betting action from a user on a specific snail in the race.
func (r *Race) placeBet(userId string, snailIdx int, amount int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	bet := UserBet{
		UserId:     userId,
		Amount:     amount,
		SnailIndex: snailIdx,
	}

	r.Snails[snailIdx].Pool += amount
	r.Pool += amount
	r.UserBets = append(r.UserBets, bet)

	database.Save(r)
	database.Save(&r.Snails[snailIdx])
}

// puntersPlaceBets lets all registered punters place their bets on the snails
// in the race.
func (r *Race) puntersPlaceBets(punters []Punter) {
	r.mu.Lock()
	defer r.mu.Unlock()

	snails := make([]Snail, 0)
	for _, s := range r.Snails {
		snails = append(snails, *s.snail)
	}

	for _, p := range punters {
		index, amount := p.GetBet(snails)
		r.Snails[index].Pool += amount
		r.Pool += amount
	}

	// Update odds for each snail after all bets are placed.
	for _, s := range r.Snails {
		odds := calculateOdds(r.Pool, s.Pool)
		s.snail.odds = odds
	}

	database.Save(r)
	for _, s := range r.Snails {
		database.Save(&s)
	}
}
