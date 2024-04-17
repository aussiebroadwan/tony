package blackjack

import "gorm.io/gorm"

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
type AchievementCallback func(userId string, achievementName string)

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

	AchievedFirstWin            bool
	Achieved100Games            bool
	Achieved3Blackjacks         bool
	Achieved1kTotalWinnings     bool
	AchievedLoss1kInOneRound    bool
	Achieved7LossesInARowThenBJ bool
	Achieved21In2Cards21Times   bool
	Achieved21In7CardsWin       bool
}

var database *gorm.DB = nil

// SetupAchievementDB initialises the database connection for storing
// achievements.
func SetupAchievementDB(db *gorm.DB) error {
	database = db
	return database.AutoMigrate(&UserAchievements{})
}

func emptyAchievementStatus() UserAchievements {
	return UserAchievements{
		RoundsPlayed: 0,

		LastShoeId:  "",
		ShoesPlayed: 0,

		RoundsWon:  0,
		RoundsLost: 0,

		RoundsSinceLastWin:             0,
		RoundsSinceLastWinBlackjack:    0,
		RoundsSinceLastWinNonBlackjack: 0,

		TotalWinnings: 0,
		TotalLosses:   0,

		NumberOfBlackjacks: 0,
		BlackjackStreak:    0,

		AchievedFirstWin:            false,
		Achieved100Games:            false,
		Achieved3Blackjacks:         false,
		Achieved1kTotalWinnings:     false,
		AchievedLoss1kInOneRound:    false,
		Achieved7LossesInARowThenBJ: false,
		Achieved21In2Cards21Times:   false,
		Achieved21In7CardsWin:       false,
	}
}

// getAchievementStatus retrieves or initialises the achievement status of a
// specific user. This is useful for seeing what has changed.
func getAchievementStatus(userId string) (UserAchievements, error) {
	achivement := UserAchievements{}
	err := database.Where("user_id = ?", userId).Attrs(emptyAchievementStatus()).FirstOrCreate(&achivement).Error
	if err != nil {
		return emptyAchievementStatus(), err
	}
	achivement.UserId = userId // Ensure the user ID is set

	return achivement, nil
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

	achievement, err := getAchievementStatus(user.Id)
	if err != nil {
		return
	}

	updatedAchievement := updateAchievementStatus(achievement, user, state)

	if checkFirstWin(achievement, updatedAchievement) {
		updatedAchievement.AchievedFirstWin = true
		callback(user.Id, FirstTimeWinner)
	}

	if check100Games(achievement, updatedAchievement) {
		updatedAchievement.Achieved100Games = true
		callback(user.Id, VeteranPlayer)
	}

	if check3Blackjacks(achievement, updatedAchievement) {
		updatedAchievement.Achieved3Blackjacks = true
		callback(user.Id, BlackjackStreak)
	}

	if checkHighRoller(achievement, updatedAchievement) {
		updatedAchievement.Achieved1kTotalWinnings = true
		callback(user.Id, HighRoller)
	}

	if checkOhShit(achievement, user) {
		updatedAchievement.AchievedLoss1kInOneRound = true
		callback(user.Id, OhShit)
	}

	if checkCombackKing(achievement, user) {
		updatedAchievement.Achieved7LossesInARowThenBJ = true
		callback(user.Id, CombackKing)
	}

	if checkPerfect21(achievement, updatedAchievement, user) {
		updatedAchievement.Achieved21In2Cards21Times = true
		callback(user.Id, Perfect21)
	}

	if checkLuckySeven(achievement, user) {
		updatedAchievement.Achieved21In7CardsWin = true
		callback(user.Id, LuckySeven)
	}

	database.Save(&updatedAchievement)
}

func checkFirstWin(before, after UserAchievements) bool {
	return !before.AchievedFirstWin && before.RoundsWon == 0 && after.RoundsWon > 0
}

func check100Games(before, after UserAchievements) bool {
	return !before.Achieved100Games && before.RoundsPlayed < 100 && after.RoundsPlayed >= 100
}

func check3Blackjacks(before, after UserAchievements) bool {
	return !before.Achieved3Blackjacks && after.BlackjackStreak >= 3
}

func checkHighRoller(before, after UserAchievements) bool {
	return !before.Achieved1kTotalWinnings && (after.TotalWinnings-after.TotalLosses) >= 1000
}

func checkOhShit(before UserAchievements, user User) bool {
	return !before.AchievedLoss1kInOneRound && user.Bet == 0 && user.InitialBet >= 1000
}

func checkCombackKing(before UserAchievements, user User) bool {
	return !before.Achieved7LossesInARowThenBJ && before.RoundsSinceLastWin >= 7 && user.Blackjack
}

func checkPerfect21(before, after UserAchievements, user User) bool {
	return !before.Achieved21In2Cards21Times && user.Blackjack && after.NumberOfBlackjacks >= 21
}

func checkLuckySeven(before UserAchievements, user User) bool {
	return !before.Achieved21In7CardsWin && user.Hand.Score() == MaximumHandScore && len(user.Hand) == 7
}
