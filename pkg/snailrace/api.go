package snailrace

import (
	"time"
)

const (
	JoinDelay  time.Duration = 30 * time.Second
	StartDelay time.Duration = 30 * time.Second

	PuntersPerRace int = 64
)

// HostRace initialises and starts a new race with given parameters. It requires
// a state change callback, an achievement callback, a message ID, and a channel
// ID to properly configure the race. If any parameter is invalid, it returns an
// error.
func HostRace(stateCb StateChangeCallback, achievementCb AchievementCallback, messageId, channelId string) error {
	if stateCb == nil || messageId == "" || channelId == "" {
		return ErrInvalidParameters
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	now := time.Now()
	race := newRace() // Simplify race creation with a safe newRace function
	race.StartAt = now.Add(StartDelay).Add(JoinDelay)

	r := &RaceState{
		Race:          race,
		State:         StateJoining,
		Step:          0,
		Snails:        make([]*Snail, 0),
		MessageId:     messageId,
		ChannelId:     channelId,
		stateCb:       stateCb,
		achievementCb: achievementCb,
	}

	manager.races[race.Id] = r
	go r.Start(now.Add(JoinDelay)) // Start the race after the joining delay

	return nil
}

// GetSnails retrieves all snails owned by a user from the database. If the user
// does not own any snails, a new snail is generated, saved to the database, and
// returned. It returns a slice of Snail objects or an error if the database
// query fails.
func GetSnails(userId string) ([]Snail, error) {
	var snails []Snail
	if err := database.Where(Snail{OwnerId: userId}).Find(&snails).Error; err != nil {
		return nil, err
	}

	if len(snails) == 0 {
		snail := GenerateSnail()
		snail.OwnerId = userId
		if err := database.Create(&snail).Error; err != nil {
			return nil, err
		}
		snails = append(snails, snail)
	}

	return snails, nil
}

// JoinRace allows a user to add a snail to a race during the joining phase. The
// function requires a user ID, a race ID, and a snail ID. It checks the
// ownership of the snail and the current state of the race. If the race is not
// in the joining state, or the snail does not belong to the user, it returns
// an error.
func JoinRace(userId, raceId, snailId string) error {
	snail := Snail{}
	if err := database.First(&snail, Snail{Id: snailId}).Error; err != nil {
		return ErrSnailNotFound
	}

	if snail.OwnerId != userId {
		return ErrNotSnailOwner
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()
	r, ok := manager.races[raceId]
	if !ok {
		return ErrRaceNotFound
	}

	if r.State != StateJoining {
		return ErrInvalidRaceState
	}

	r.Join(snail)
	r.stateCb(*r, r.MessageId, r.ChannelId)
	return nil
}

// PlaceBet allows a user to place a bet on a snail in a specific race. The
// function requires a user ID, a race ID, the index of the snail within the
// race, and the bet amount. It checks if the race is in the betting state and
// if the race exists. It returns an error if the race is not found or not in
// the correct state.
func PlaceBet(userId, raceId string, snailIdx int, amount int64) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	r, ok := manager.races[raceId]
	if !ok {
		return ErrRaceNotFound
	}

	if r.State != StateBetting {
		return ErrInvalidRaceState
	}

	r.Race.placeBet(userId, snailIdx, amount)
	return nil
}
