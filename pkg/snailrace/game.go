package snailrace

import "gorm.io/gorm"

const (
	NUMBER_OF_PUNTERS = 256
)

var database *gorm.DB = nil

// SetupSnailraceDB sets up the database for the snailrace game.
func SetupSnailraceDB(db *gorm.DB) error {
	database = db

	database.AutoMigrate(&Punter{}, &Snail{}, &Race{}, &SnailRaceLink{})

	setupPunters()

	return nil

}

func setupPunters() {
	// Get all punters from the database
	var punters []Punter
	database.Find(&punters)

	// If there are no punters, generate them
	for len(punters) < NUMBER_OF_PUNTERS {
		p := GeneratePunter()
		database.Create(&p)
		punters = append(punters, p)
	}
}
