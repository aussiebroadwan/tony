package snailrace

import (
	"math/rand"
	"time"
)

const (
	MinRacesPerDay = 7
	MaxRacesPerDay = 12

	MinSnailsPerRace = 7
	MaxSnailsPerRace = 12

	RaceReadyDuration = time.Minute * 5
)

// weekScheduleHours maps each day of the week to the operating hours for races.
var weekScheduleHours = map[time.Weekday][2]time.Duration{
	time.Monday:    {11 * time.Hour, 20 * time.Hour},
	time.Tuesday:   {11 * time.Hour, 20 * time.Hour},
	time.Wednesday: {11 * time.Hour, 20 * time.Hour},
	time.Thursday:  {11 * time.Hour, 20 * time.Hour},
	time.Friday:    {10 * time.Hour, 22 * time.Hour},
	time.Saturday:  {10 * time.Hour, 22 * time.Hour},
	time.Sunday:    {9 * time.Hour, 21 * time.Hour},
}

type RaceReadyCallback func() (stateCb StateChangeCallback, achievementCb AchievementCallback, messageId, channelId string)

// LaunchSnailraceTV initializes the race broadcasting system, scheduling and
// managing races continuously. This must be called in a goroutine by an
// application OnMount() function as it should only run once
func LaunchSnailraceTV(onRaceReady RaceReadyCallback) {

	for {
		ticker := time.NewTicker(time.Second)

		today := time.Now().Format(time.DateOnly)
		races := getTodaysScheduledRaces()
		raceIdx := 0

		for range ticker.C {
			if today != time.Now().Format(time.DateOnly) {
				ticker.Stop()
				break
			}

			raceIdx = manageRaces(races, raceIdx, onRaceReady)
		}
	}
}

// manageRaces checks for upcoming races and triggers them at the scheduled
// start times.
func manageRaces(races []*Race, raceIdx int, onRaceReady RaceReadyCallback) int {
	if raceIdx >= len(races) {
		return raceIdx
	}

	targetTime := races[raceIdx].StartAt.Add(-RaceReadyDuration)

	if time.Now().After(targetTime) {
		targetTime := races[raceIdx].StartAt.Add(-RaceReadyDuration)
		if time.Now().After(targetTime) {
			stateCb, achievementCb, messageId, channelId := onRaceReady()
			state := updateScheduledRaceState(races[raceIdx], stateCb, achievementCb, messageId, channelId)
			go state.Start(time.Time{})
			return raceIdx + 1
		}
	}

	return raceIdx
}

// updateScheduledRaceState updates the state of a scheduled race with the
// provided callback values to start rendering the race.
func updateScheduledRaceState(race *Race, stateCb StateChangeCallback, achievementCb AchievementCallback, messageId, channelId string) *RaceState {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	state := manager.races[race.Id]
	state.stateCb = stateCb
	state.achievementCb = achievementCb
	state.MessageId = messageId
	state.ChannelId = channelId
	manager.races[race.Id] = state
	return state
}

// getTodaysRaces retrieves all races scheduled for today from the database.
func getTodaysScheduledRaces() []*Race {
	var races []Race
	if err := database.Where(Race{UserHosted: false}).Find(&races).Error; err != nil || len(races) == 0 {
		return []*Race{}
	}

	todaysRaces := filterRacesByDate(races, time.Now().Format(time.DateOnly))
	if len(todaysRaces) == 0 {
		generateTodaysScheduledRaces(todaysRaces)
	}

	return todaysRaces
}

// filterRacesByDate filters races to find those that occur on the specified
// date.
func filterRacesByDate(races []Race, date string) []*Race {
	var todaysRaces []*Race
	for idx := range races {
		if races[idx].StartAt.Format("2006-01-02") == date {
			todaysRaces = append(todaysRaces, &races[idx])
		}
	}
	return todaysRaces
}

// generateTodaysRaces creates races for today based on the weekly schedule and
// adds them to the database.
func generateTodaysScheduledRaces(races []*Race) {
	todayTime, _ := time.Parse(time.DateOnly, time.Now().Format(time.DateOnly))
	twoRaceTimes := weekScheduleHours[time.Now().Weekday()]

	// If the current time is after the first schduled race then skip the
	// races for today and start over tomorrow. This is so if the bot is
	// launched in the afternoon it wont schdule an entire day of races
	if time.Now().After(todayTime.Add(twoRaceTimes[0])) {
		return
	}

	// Calculate the time between todays races
	raceTimeRange := twoRaceTimes[1] - twoRaceTimes[0]
	numRaces := MinRacesPerDay + rand.Intn(MaxRacesPerDay-MinRacesPerDay)
	spreadTime := raceTimeRange / time.Duration(numRaces)

	// Get all snails that are eligible to participate today
	snails := getEligibleScheduledSnails()
	if len(snails) == 0 {
		return
	}
	snailsUsed := 0

	// Create and schedule todays races
	for i := 0; i < numRaces; i++ {
		startTime := twoRaceTimes[0] + spreadTime*time.Duration(i)
		races[i] = newRace(todayTime.Add(startTime), false)
		snailsUsed += manageNewScheduledRaceState(races[i], snails[snailsUsed:])
	}
}

// getEligibleScheduledSnails retrieves all snails that are eligible to
// participate in a scheduled races. This also shuffles the snails to
// randomise the order.
func getEligibleScheduledSnails() []Snail {
	var snails []Snail
	err := database.Where(Snail{OwnerId: GeneratedOwnerId}).Find(&snails).Error
	if err != nil {
		return make([]Snail, 0)
	}
	rand.Shuffle(len(snails), func(i, j int) { snails[i], snails[j] = snails[j], snails[i] })
	return snails
}

// manageNewScheduledRaceState creates a new state instance to the scheduled
// race. It will also add a random number of snails to the race and setup the
// odds ready for the Betting stage.
func manageNewScheduledRaceState(race *Race, snails []Snail) int {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	state := &RaceState{
		Race:           race,
		State:          StateBetting, // Start in the betting stage to not let user's join
		Step:           0,
		Snails:         make([]*Snail, 0),
		snailsToRemove: make([]string, 0),
	}

	// Add a snail to the race
	numSnails := MinSnailsPerRace + rand.Intn(MaxSnailsPerRace-MinSnailsPerRace)
	for j := 0; j < numSnails; j++ {
		state.Join(snails[j])
	}
	state.puntersPlaceBets()

	// Add race to the manager
	manager.races[race.Id] = state
	return numSnails
}
