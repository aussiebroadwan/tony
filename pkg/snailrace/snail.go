package snailrace

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
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

func SnailType(t int) string {
	switch t {
	case Prisimshell:
		return "Prisimshell"
	case Thunderhorn:
		return "Thunderhorn"
	case Royalcrest:
		return "Royalcrest"
	case Circuitshell:
		return "Circuitshell"
	case Infernoshell:
		return "Infernoshell"
	case Obsidianshell:
		return "Obsidianshell"
	case Chillshell:
		return "Chillshell"
	case Stoneclad:
		return "Stoneclad"
	default:
		return "Unknown"
	}
}

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

var random = rand.New(rand.NewSource(time.Now().Unix()))

func hashSnail(s Snail, tag string) []byte {
	h := md5.New()
	io.WriteString(h, tag)
	io.WriteString(h, fmt.Sprintf("%f", s.Speed))
	io.WriteString(h, fmt.Sprintf("%f", s.Acceleration))
	io.WriteString(h, fmt.Sprintf("%f", s.Weight))
	io.WriteString(h, fmt.Sprintf("%d", s.Stamina))
	io.WriteString(h, fmt.Sprintf("%f", s.Luck))
	return h.Sum(nil)
}

// GenerateSnail generates a random snail with random stats and a random type.
func GenerateSnail() Snail {

	s := Snail{
		Name:    generateSnailName(random),
		OwnerId: "",
		Type:    random.Intn(8), // 8 snail types

		Speed:        random.Float64()*9 + 1,
		Acceleration: random.Float64()*0.9 + 0.1,
		Weight:       random.Float64()*9 + 1,
		Stamina:      random.Intn(91) + 10,
		Luck:         random.Float64() - 0.5,

		Prev1Place: 0,
		Prev2Place: 0,
		Prev3Place: 0,
	}

	// Hash the snail's name and stats to generate a unique ID

	s.Id = fmt.Sprintf("snail_%d%x", s.Type, hashSnail(s, s.Name))

	return s
}

// SimulateRace simulates a snail race using the attributes of the snail and a
// race ID to generate a random seed. It outputs an array of positions
// representing the snail's position for each second of the race up to 100
// seconds or until the snail crosses the finish line at position 100.
func (s Snail) SimulateRace(raceID string) []float64 {
	// Generate a random seed based on the race ID
	hash := int64(binary.BigEndian.Uint64(hashSnail(s, raceID)))
	r := rand.New(rand.NewSource(hash))

	positions := make([]float64, 0)
	positions = append(positions, 0)
	currentSpeed := 0.0
	time := 0

	for positions[len(positions)-1] < 100 && time < 100 {
		// Increase speed based on acceleration
		if currentSpeed < s.Speed {
			maxSpeed := s.Speed
			if time > s.Stamina {
				maxSpeed = float64(s.Stamina) / s.Weight
			}

			currentSpeed += min(s.Acceleration, maxSpeed-currentSpeed)
		}

		// Apply luck and random fluctuations
		currentSpeed += (r.Float64()*2 - 1) + (s.Luck * 0.1)

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
