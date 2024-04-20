package snailrace

import (
	"crypto/sha1"
	"fmt"
	"math/rand"
	"time"
)

const (
	// Snail types
	Prisimshell = iota
	Thunderhorn
	Royalcrest
	Circuitshell
	Infernoshell
	Obsidianshell
	Chillshell
	Stoneclad
)

type Snail struct {
	Id      string `gorm:"primaryKey"`
	OwnerId string
	Name    string
	Type    int

	// Stats
	Speed        float64 // Range 1 to 10
	Acceleration float64 // Range 0.1 to 1.0
	Weight       float64 // Range 1 to 10
	Stamina      int     // Range 10 to 100
	Luck         float64 // Range -0.5 to 0.5

	// History
	Prev1Place int
	Prev2Place int
	Prev3Place int
}

func (s *Snail) generateId() {
	// Hash the snail's name and stats to generate a unique ID
	hasher := sha1.New()
	hasher.Write([]byte(s.Name))
	hasher.Write([]byte(fmt.Sprintf("%f", s.Speed)))
	hasher.Write([]byte(fmt.Sprintf("%f", s.Acceleration)))
	hasher.Write([]byte(fmt.Sprintf("%f", s.Weight)))
	hasher.Write([]byte(fmt.Sprintf("%d", s.Stamina)))
	hasher.Write([]byte(fmt.Sprintf("%f", s.Luck)))

	s.Id = fmt.Sprintf("snail_%d%x", s.Type, hasher.Sum(nil))
}

// GenerateSnail generates a random snail with random stats and a random type.
func GenerateSnail() Snail {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	s := Snail{
		Name:    generateSnailName(r),
		OwnerId: "",
		Type:    r.Intn(8), // 8 snail types

		Speed:        r.Float64()*9 + 1,
		Acceleration: r.Float64()*0.9 + 0.1,
		Weight:       r.Float64()*9 + 1,
		Stamina:      r.Intn(91) + 10,
		Luck:         r.Float64() - 0.5,

		Prev1Place: 0,
		Prev2Place: 0,
		Prev3Place: 0,
	}
	s.generateId()

	return s
}

// SimulateRace simulates a snail race using the attributes of the snail and a
// race ID to generate a random seed. It outputs an array of positions
// representing the snail's position for each second of the race up to 100
// seconds or until the snail crosses the finish line at position 100.
func (s Snail) SimulateRace(raceID int64) []float64 {
	r := rand.New(rand.NewSource(raceID))

	positions := make([]float64, 0)
	positions = append(positions, 0)
	currentSpeed := 0.0
	time := 0

	for positions[len(positions)-1] < 100 && time < 100 {
		// Increase speed based on acceleration
		if currentSpeed < s.Speed {
			currentSpeed += min(s.Acceleration, s.Speed-currentSpeed)
		}

		// Apply weight impact
		currentSpeed -= s.Weight * 0.01

		// Apply stamina impact
		if time > s.Stamina {
			currentSpeed *= 0.9
		}

		// Apply luck and random fluctuations
		currentSpeed += (r.Float64()*2 - 1) * s.Luck

		// Ensure speed does not drop below 0
		currentSpeed = max(currentSpeed, 0)

		// Update position
		newPosition := min(100, positions[len(positions)-1]+currentSpeed)
		positions = append(positions, newPosition)

		time++
	}
	return positions
}

// hasHistoricalPlacement checks if a snail has historical placements in the
// top 3.
func (snail Snail) hasHistoricalPlacement() bool {
	return (snail.Prev1Place <= 3 && snail.Prev1Place > 0) || (snail.Prev2Place <= 3 && snail.Prev2Place > 0)
}
