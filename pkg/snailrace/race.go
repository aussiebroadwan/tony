package snailrace

import (
	"sync"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Race struct {
	Id string `gorm:"primaryKey"`

	Pool int64 // Total amount of money in the pool from the Punters

	mu       sync.Mutex      `gorm:"-"`
	Snails   []SnailRaceLink `gorm:"-"`
	UserBets []UserBet       `gorm:"-"`
}

type UserBet struct {
	UserId     string
	Amount     int64
	SnailIndex int
}

type SnailRaceLink struct {
	gorm.Model

	RaceId  string
	SnailId string
	snail   Snail `gorm:"-"`

	Pool int64 // Total amount of money in the pool for this Snail from the Punters
}

func calculateOdds(racePool, snailPool int64) float64 {
	return float64(snailPool) / float64(racePool)
}

func newRace(existingRaces map[string]*Race) *Race {
	// Generate Unique ID
	id := uuid.New().String()[24:]
	_, ok := existingRaces[id]
	for ok {
		id = uuid.New().String()[24:]
		_, ok = existingRaces[id]
	}

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

func (r *Race) joinRace(s Snail) {
	r.mu.Lock()
	defer r.mu.Unlock()

	link := SnailRaceLink{
		RaceId:  r.Id,
		SnailId: s.Id,
		snail:   s,
		Pool:    0,
	}
	database.Create(&link)
	r.Snails = append(r.Snails, link)
}

func (r *Race) placeBet(userId string, snailIdx int, amount int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Snails[snailIdx].Pool += amount
	r.Pool += amount
	r.UserBets = append(r.UserBets, UserBet{
		UserId:     userId,
		Amount:     amount,
		SnailIndex: snailIdx,
	})
	database.Save(r)
	database.Save(&r.Snails[snailIdx])
}

func (r *Race) puntersPlaceBets(punters []Punter) {
	r.mu.Lock()
	defer r.mu.Unlock()

	snails := make([]Snail, 0)
	for _, s := range r.Snails {
		snails = append(snails, s.snail)
	}

	for _, p := range punters {
		index, amount := p.GetBet(snails)
		r.Snails[index].Pool += amount
		r.Pool += amount
	}

	database.Save(r)
	for _, s := range r.Snails {
		database.Save(&s)
	}
}
