package snailrace

import (
	"sync"
	"time"

	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"
)

const (
	NumberOfPunters = 256
)

var (
	database *gorm.DB
	manager  *RaceManager
)

// setupPunters ensures the required number of punters are in the database.
func setupPunters() error {
	var count int64
	result := database.Model(&Punter{}).Count(&count)
	if result.Error != nil {
		return result.Error
	}

	// Generate and save missing punters
	for i := int(count); i < NumberOfPunters; i++ {
		punter := GeneratePunter()
		if err := database.Create(&punter).Error; err != nil {
			return err
		}
		log.WithField("src", "snailrace").WithField("punter", punter.ID).Info("Generated punter")
	}

	return nil
}

// RaceManager manages the races and their states.
type RaceManager struct {
	races map[string]*RaceState
	mu    sync.Mutex
}

// InitializeRaceManager sets up a new game state with race manager.
func InitializeRaceManager() {
	manager = &RaceManager{
		races: make(map[string]*RaceState),
	}
	go manager.raceCleaner()
}

// raceCleaner periodically removes finished races from the manager.
func (rm *RaceManager) raceCleaner() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rm.cleanupRaces()
	}
}

// cleanupRaces removes races that are finished.
func (rm *RaceManager) cleanupRaces() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for id, state := range rm.races {
		if state.State == StateFinished {
			delete(rm.races, id)
		}
	}
}

// SetupSnailraceDB initializes the database and sets up the environment for the snail racing game.
func SetupSnailraceDB(db *gorm.DB) error {
	database = db

	if err := database.AutoMigrate(&Punter{}, &Snail{}, &Race{}, &SnailRaceLink{}); err != nil {
		return err
	}

	if err := setupPunters(); err != nil {
		return err
	}

	InitializeRaceManager()
	return nil
}
