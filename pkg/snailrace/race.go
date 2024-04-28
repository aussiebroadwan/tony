package snailrace

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"
)

const (
	// Bet types
	BetTypeWin      int = 0
	BetTypePlace    int = 1
	BetTypeEachWay  int = 2
	BetTypeQuinella int = 3
	BetTypeExacta   int = 4
	BetTypeTrifecta int = 5
)

// Race struct represents a single racing event.
type Race struct {
	ID string `gorm:"primaryKey"`

	UserHosted bool
	StartAt    time.Time // Time when the race is scheduled to start
	Pool       int64     // Total amount of money in the pool from the Punters

	SnailRaceLinks []SnailRaceLink
	UserBets       []UserBet

	mu sync.Mutex `gorm:"-"`
}

// UserBet struct represents a bet made by a user on a snail.
type UserBet struct {
	gorm.Model

	RaceID string

	UserId      string
	Amount      int64
	Type        int
	Snail1Index int
	Snail2Index int
	Snail3Index int
}

// SnailRaceLink struct links a snail to a race with a specific betting pool.
type SnailRaceLink struct {
	gorm.Model

	RaceID string

	SnailID string
	Snail   Snail

	Pool int64 // Total amount of money in the pool for this Snail from the Punters
}

// CalculateOdds computes the betting odds for a snail based on the total pool
// and snail's pool.
func CalculateOdds(racePool, snailPool int64) float64 {
	return float64(racePool) / float64(snailPool)
}

// CalculateOdds computes the betting odds for a snail based on the total pool
// and snail's pool.
func CalculatePlaceOdds(racePool, snailPool int64) float64 {
	winOdds := CalculateOdds(racePool, snailPool)

	// Place odds are 1/3 of the win odds
	diff := winOdds - 1.0
	return 1.0 + diff/3.0
}

// newRace creates and initialises a new race with a unique identifier.
func newRace(startTime time.Time, hosted bool) *Race {
	id := uuid.New().String()[24:] // Truncating UUID to get a shorter ID.

	r := &Race{
		ID:             id,
		Pool:           0,
		UserHosted:     hosted,
		StartAt:        startTime,
		SnailRaceLinks: make([]SnailRaceLink, 0),
		UserBets:       make([]UserBet, 0),
		mu:             sync.Mutex{},
	}
	database.Create(r)
	return r
}

// joinRace adds a snail to the race and registers it in the database.
func (r *Race) joinRace(s Snail) {
	r.mu.Lock()
	defer r.mu.Unlock()

	link := SnailRaceLink{
		RaceID: r.ID,
		Snail:  s,
		Pool:   0,
	}
	database.Create(&link) // Register the new link in the database.
	r.SnailRaceLinks = append(r.SnailRaceLinks, link)
}

// placeBet records a betting action from a user on a specific snail in the race.
func (r *Race) placeBet(userId string, amount int64, betType int, snailIdx ...int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	bet := UserBet{
		RaceID:      r.ID,
		UserId:      userId,
		Amount:      amount,
		Type:        betType,
		Snail1Index: 0,
		Snail2Index: 0,
		Snail3Index: 0,
	}

	switch betType {
	case BetTypeWin, BetTypePlace, BetTypeEachWay:
		if len(snailIdx) != 1 {
			log.WithField("src", "snailrace").WithError(ErrInvalidBet).Error("Win, Place and Eachway bets requires a single snail index")
			return ErrInvalidBet
		}
		bet.Snail1Index = snailIdx[0]
		r.SnailRaceLinks[snailIdx[0]].Pool += amount
		database.Save(&r.SnailRaceLinks[snailIdx[0]])
	case BetTypeQuinella, BetTypeExacta:
		if len(snailIdx) != 2 {
			log.WithField("src", "snailrace").WithError(ErrInvalidBet).Error("Quinella and Exacta bets requires two snail indexes")
			return ErrInvalidBet
		}
		bet.Snail1Index = snailIdx[0]
		bet.Snail2Index = snailIdx[1]

		perSnail := float64(amount) / 2.0

		r.SnailRaceLinks[snailIdx[0]].Pool += int64(perSnail)
		r.SnailRaceLinks[snailIdx[1]].Pool += int64(perSnail)

		database.Save(&r.SnailRaceLinks[snailIdx[0]])
		database.Save(&r.SnailRaceLinks[snailIdx[1]])
	case BetTypeTrifecta:
		if len(snailIdx) != 3 {
			log.WithField("src", "snailrace").WithError(ErrInvalidBet).Error("Trifecta bets requires three snail indexes")
			return ErrInvalidBet
		}
		bet.Snail1Index = snailIdx[0]
		bet.Snail2Index = snailIdx[1]
		bet.Snail3Index = snailIdx[2]

		perSnail := float64(amount) / 3.0

		r.SnailRaceLinks[snailIdx[0]].Pool += int64(perSnail)
		r.SnailRaceLinks[snailIdx[1]].Pool += int64(perSnail)
		r.SnailRaceLinks[snailIdx[2]].Pool += int64(perSnail)

		database.Save(&r.SnailRaceLinks[snailIdx[0]])
		database.Save(&r.SnailRaceLinks[snailIdx[1]])
		database.Save(&r.SnailRaceLinks[snailIdx[2]])
	default:
		log.WithField("src", "snailrace").WithError(ErrInvalidBet).Error("Invalid bet type")
		return ErrInvalidBet
	}

	// Save the bet to the database
	// database.Save(&bet)

	r.Pool += amount
	r.UserBets = append(r.UserBets, bet)
	database.Save(r)
	return nil
}

// puntersPlaceBets lets all registered punters place their bets on the snails
// in the race.
func (r *Race) puntersPlaceBets(punters []Punter) {
	r.mu.Lock()
	defer r.mu.Unlock()

	snails := make([]Snail, 0)
	for _, s := range r.SnailRaceLinks {
		snails = append(snails, s.Snail)
	}

	for _, p := range punters {
		index, amount := p.GetBet(snails)
		r.SnailRaceLinks[index].Pool += amount
		r.Pool += amount
	}
	log.WithField("src", "snailrace").WithField("race", r.ID).Infof("Total pool: %d", r.Pool)

	database.Save(r)
	for _, s := range r.SnailRaceLinks {
		odds := CalculateOdds(r.Pool, s.Pool)
		log.WithField("src", "snailrace").WithField("race", r.ID).WithField("snail", s.Snail.ID).WithField("odds", odds).Info("Updated snail odds")
		database.Save(&s)
	}
}
