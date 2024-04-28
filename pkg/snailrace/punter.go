package snailrace

import (
	"math/rand"
	"sort"

	"gorm.io/gorm"
)

const (
	PunterMaxBudget = 150
	PunterMinBudget = 15
)

const (
	// Punter Preferences
	PreferencePrisimshell uint16 = 1 << iota
	PreferenceThunderhorn
	PreferenceRoyalcrest
	PreferenceCircuitshell
	PreferenceInfernoshell
	PreferenceObsidianshell
	PreferenceChillshell
	PreferenceStoneclad
	PreferenceSnailSpeed
	PreferenceSnailAcceleration
	PreferenceSnailWeight
	PreferenceSnailStamina
	PreferenceSnailLuck
	PreferenceHistorical1
	PreferenceHistorical2
)

// randomPreferences generates a random set of preferences as a bitmask.
func randomPreferences() uint16 {
	return uint16(rand.Float64() * 65535)
}

// Punter represents a participant with a budget and preferences.
type Punter struct {
	gorm.Model

	Budget     int64 // How much can the punter bet
	Preference uint16
}

// GeneratePunter creates a new punter with a default budget and random preferences.
func GeneratePunter() Punter {
	return Punter{
		Budget:     int64(rand.Intn(PunterMaxBudget-PunterMinBudget) + PunterMinBudget),
		Preference: randomPreferences(),
	}
}

// getHighestAttribute retrieves the snail ID with the highest specific attribute.
func getHighestAttribute(snails []Snail, attribute string) string {
	sort.Slice(snails, func(i, j int) bool {
		switch attribute {
		case "Speed":
			return snails[i].Speed > snails[j].Speed
		case "Acceleration":
			return snails[i].Acceleration > snails[j].Acceleration
		case "Weight":
			return snails[i].Weight > snails[j].Weight
		case "Stamina":
			return snails[i].Stamina > snails[j].Stamina
		case "Luck":
			return snails[i].Luck > snails[j].Luck
		default:
			return false
		}
	})
	return snails[0].ID
}

// GetBet determines which snail a punter should bet on based on their
// preferences and ramdomness.
func (p Punter) GetBet(snails []Snail) (index int, amount int64) {

	possiblePicks := make([]int, 0)
	attributes := []string{"Speed", "Acceleration", "Weight", "Stamina", "Luck"}
	highestAttributes := make(map[string]string)

	// Precalc the snails with the highest attribute for each attribute
	for _, attr := range attributes {
		highestAttributes[attr] = getHighestAttribute(snails, attr)
	}

	// Check if any snails match the punter's preferences
	for i, snail := range snails {
		if p.Preference&1<<snail.Type != 0 {
			possiblePicks = append(possiblePicks, i)
		}

		// Check if the snail has the highest attribute for each attribute
		for a, attr := range attributes {
			preferenceFlag := PreferenceSnailSpeed << a
			if snail.ID == highestAttributes[attr] && p.Preference&preferenceFlag != 0 {
				possiblePicks = append(possiblePicks, i)
			}
		}

		// Check if the snail has historical placements
		if snail.hasHistoricalPlacement() && p.Preference&(PreferenceHistorical1|PreferenceHistorical2) != 0 {
			possiblePicks = append(possiblePicks, i)
		}
	}

	// If there are possible picks, pick one at random
	if len(possiblePicks) > 0 {
		return possiblePicks[rand.Intn(len(possiblePicks))], p.Budget
	}

	// If there are no possible picks, pick a random snail
	return rand.Intn(len(snails)), p.Budget
}
