package blackjack

import (
	"gorm.io/gorm"
)

// Constants for the names of various achievements.
const (
	FirstTimeWinner = "blackjack_1st_win"
	VeteranPlayer   = "blackjack_100_games"
	BlackjackStreak = "blackjack_3_bjs"
	HighRoller      = "blackjack_1k_total_winnings"
	OhShit          = "blackjack_loss_1k_in_one_round"
	CombackKing     = "blackjack_7_losses_in_a_row_then_bj"
	Perfect21       = "blackjack_21_in_2_cards_21_times"
	LuckySeven      = "blackjack_21_in_7_cards_win"
)

// AchievementCallback defines a callback function type for when an achievement
// is unlocked. Important as we dont want this package to depend on any other
// package.
type AchievementCallback func(userId string, achievementName string) bool

// These are mostly used for checking the conditions of achievements and
// shouldn't be used outside of this package.
type achievementChecker func(before, after UserAchievements, user User) bool
type achievementMarker func(achievementName string, checker achievementChecker) bool

// UserAchievements tracks the achievements of a user in blackjack.
type UserAchievements struct {
	UserId string `gorm:"primaryKey"`

	RoundsPlayed int64

	LastShoeId  string
	ShoesPlayed int64

	RoundsWon  int64
	RoundsLost int64

	RoundsSinceLastWin             int64
	RoundsSinceLastWinBlackjack    int64
	RoundsSinceLastWinNonBlackjack int64

	TotalWinnings int64
	TotalLosses   int64

	NumberOfBlackjacks int // 21 in 2 cards
	BlackjackStreak    int // Number of blackjacks in a row

	// Achievement flags
	AchievedFirstWin            bool
	Achieved100Games            bool
	Achieved3Blackjacks         bool
	Achieved1kTotalWinnings     bool
	AchievedLoss1kInOneRound    bool
	Achieved7LossesInARowThenBJ bool
	Achieved21In2Cards21Times   bool
	Achieved21In7CardsWin       bool

	// Non-persistent map to track which achievements have been unlocked during the session.
	achivementsUnlocked map[string]bool `gorm:"-"`
}

// populateAchievementsUnlocked updates the map with the current state of each achievement.
func (a *UserAchievements) populateAchievementsUnlocked() {
	// Initialise the map if it is nil.
	a.achivementsUnlocked = make(map[string]bool)

	// Populate the map with the current state of each achievement.
	a.achivementsUnlocked[FirstTimeWinner] = a.AchievedFirstWin
	a.achivementsUnlocked[VeteranPlayer] = a.Achieved100Games
	a.achivementsUnlocked[BlackjackStreak] = a.Achieved3Blackjacks
	a.achivementsUnlocked[HighRoller] = a.Achieved1kTotalWinnings
	a.achivementsUnlocked[OhShit] = a.AchievedLoss1kInOneRound
	a.achivementsUnlocked[CombackKing] = a.Achieved7LossesInARowThenBJ
	a.achivementsUnlocked[Perfect21] = a.Achieved21In2Cards21Times
	a.achivementsUnlocked[LuckySeven] = a.Achieved21In7CardsWin

}

var database *gorm.DB = nil

// SetupAchievementDB initialises the database connection for storing
// achievements.
func SetupAchievementDB(db *gorm.DB) error {
	database = db
	return database.AutoMigrate(&UserAchievements{})
}

// emptyAchievementStatus creates a new UserAchievements instance with default
// values.
func emptyAchievementStatus() UserAchievements {
	return UserAchievements{
		RoundsPlayed:                   0,
		LastShoeId:                     "",
		ShoesPlayed:                    0,
		RoundsWon:                      0,
		RoundsLost:                     0,
		RoundsSinceLastWin:             0,
		RoundsSinceLastWinBlackjack:    0,
		RoundsSinceLastWinNonBlackjack: 0,
		TotalWinnings:                  0,
		TotalLosses:                    0,
		NumberOfBlackjacks:             0,
		BlackjackStreak:                0,
		AchievedFirstWin:               false,
		Achieved100Games:               false,
		Achieved3Blackjacks:            false,
		Achieved1kTotalWinnings:        false,
		AchievedLoss1kInOneRound:       false,
		Achieved7LossesInARowThenBJ:    false,
		Achieved21In2Cards21Times:      false,
		Achieved21In7CardsWin:          false,
		achivementsUnlocked:            make(map[string]bool),
	}
}

// getAchievementStatus retrieves or initialises the achievement status of a
// specific user. This is useful for seeing what has changed.
func getAchievementStatus(userId string) (UserAchievements, error) {
	achievement := UserAchievements{UserId: userId}
	err := database.Where(achievement).Attrs(emptyAchievementStatus()).FirstOrCreate(&achievement).Error
	if err != nil {
		return emptyAchievementStatus(), err
	}
	achievement.UserId = userId // Ensure the user ID is set

	// Populate the achievement map based on current database records.
	achievement.populateAchievementsUnlocked()

	return achievement, nil
}

// updateAchievementStatus updates the achievements based on the current game
// state and bet outcomes. This can be compared to the previous state to see
// what has changed.
func updateAchievementStatus(before UserAchievements, user User, state GameState) UserAchievements {
	after := before

	after.RoundsPlayed++
	if user.Bet > 0 && user.Bet > user.InitialBet {
		// Player won
		after.RoundsWon++
		after.TotalWinnings += user.Bet
		after.RoundsSinceLastWin = 0

		if user.Blackjack {
			// Update blackjack streak
			after.NumberOfBlackjacks++
			after.RoundsSinceLastWinBlackjack = 0
			after.BlackjackStreak++
		} else {
			// Non blackjack win
			after.RoundsSinceLastWinNonBlackjack = 0
			after.BlackjackStreak = 0
		}
	} else if user.Bet == 0 {
		// Player lost
		after.RoundsLost++
		after.TotalLosses += user.InitialBet
	}

	// Update Shoe information
	if state.Id != before.LastShoeId {
		after.LastShoeId = state.Id
		after.ShoesPlayed++
	}

	return after
}

// UpdateAchievementProgress handles the logic for checking and updating user
// achievements. It will call the callback function for each achievement that is
// unlocked.
func UpdateAchievementProgress(user User, state GameState, callback AchievementCallback) {
	if database == nil {
		return
	}

	// Get the current achievement status for the user.
	before, err := getAchievementStatus(user.Id)
	if err != nil {
		return
	}

	// Update the achievement status based on the current game state.
	after := updateAchievementStatus(before, user, state)
	after.processAchievements(user, before, callback)

	database.Save(&after)
}

// ProcessAchievements checks and marks achievements based on game progress.
func (a *UserAchievements) processAchievements(user User, before UserAchievements, callback AchievementCallback) {
	markerFunc := buildAchievementMarker(user, before, *a, callback)

	// List of achievement checks. Extend this pattern for other achievements.
	a.AchievedFirstWin = markerFunc(FirstTimeWinner, checkFirstWin)
	a.Achieved100Games = markerFunc(VeteranPlayer, check100Games)
	a.Achieved3Blackjacks = markerFunc(BlackjackStreak, check3Blackjacks)
	a.Achieved1kTotalWinnings = markerFunc(HighRoller, checkHighRoller)
	a.AchievedLoss1kInOneRound = markerFunc(OhShit, checkOhShit)
	a.Achieved7LossesInARowThenBJ = markerFunc(CombackKing, checkCombackKing)
	a.Achieved21In2Cards21Times = markerFunc(Perfect21, checkPerfect21)
	a.Achieved21In7CardsWin = markerFunc(LuckySeven, checkLuckySeven)
}

// buildAchievementMarker creates a closure that applies a given checker
// function to determine if an achievement should be marked.
func buildAchievementMarker(user User, before, after UserAchievements, callback AchievementCallback) achievementMarker {
	return func(achievementName string, checker achievementChecker) bool {
		if before.achivementsUnlocked[achievementName] {
			return true
		}

		if !checker(before, after, user) {
			return false
		}

		return callback(user.Id, achievementName)
	}
}

func checkFirstWin(before, after UserAchievements, user User) bool {
	return !before.AchievedFirstWin && before.RoundsWon == 0 && after.RoundsWon > 0
}

func check100Games(before, after UserAchievements, user User) bool {
	return !before.Achieved100Games && before.RoundsPlayed < 100 && after.RoundsPlayed >= 100
}

func check3Blackjacks(before, after UserAchievements, user User) bool {
	return !before.Achieved3Blackjacks && after.BlackjackStreak >= 3
}

func checkHighRoller(before, after UserAchievements, user User) bool {
	return !before.Achieved1kTotalWinnings && (after.TotalWinnings-after.TotalLosses) >= 1000
}

func checkOhShit(before, after UserAchievements, user User) bool {
	return !before.AchievedLoss1kInOneRound && user.Bet == 0 && user.InitialBet >= 1000
}

func checkCombackKing(before, after UserAchievements, user User) bool {
	return !before.Achieved7LossesInARowThenBJ && before.RoundsSinceLastWin >= 7 && user.Blackjack
}

func checkPerfect21(before, after UserAchievements, user User) bool {
	return !before.Achieved21In2Cards21Times && user.Blackjack && after.NumberOfBlackjacks >= 21
}

func checkLuckySeven(before, after UserAchievements, user User) bool {
	return !before.Achieved21In7CardsWin && user.Hand.Score() == MaximumHandScore && len(user.Hand) == 7
}
