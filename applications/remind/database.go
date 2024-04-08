package remind

import (
	"time"

	"github.com/bwmarrin/discordgo"

	log "github.com/sirupsen/logrus"

	_ "gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func SetupRemindersDB(db *gorm.DB, session *discordgo.Session) {
	// AutoMigrate will create or update the reminders table to match the Reminder struct
	if err := db.AutoMigrate(&Reminder{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Load all reminders from the database
	reminders, err := LoadReminders(db)
	if err != nil {
		log.Fatalf("Failed to load reminders: %v", err)
	}

	// Iterate over each reminder
	for _, reminder := range reminders {
		// If the reminder has already been reminded, skip it
		if reminder.Reminded {
			continue
		}

		// Add the reminder to scheduler
		Add(reminder.ID, reminder)
	}
}

func LoadReminders(db *gorm.DB) ([]Reminder, error) {
	var reminders []Reminder
	result := db.Where("reminded = ?", false).Find(&reminders)
	return reminders, result.Error
}

func AddReminder(db *gorm.DB, createdBy string, triggerTime time.Time, session *discordgo.Session, channelId string, message string) (uint, error) {
	reminder := Reminder{
		CreatedBy:   createdBy,
		ChannelID:   channelId,
		TriggerTime: triggerTime,
		Message:     message,
		Reminded:    false,
	}

	result := db.Create(&reminder)
	if result.Error != nil {
		return 0, result.Error
	}

	// Add the reminder to the reminders package
	Add(reminder.ID, reminder)
	return reminder.ID, nil
}

func DeleteReminder(db *gorm.DB, id uint, user string) error {
	result := db.Delete(&Reminder{}, id)
	if result.Error != nil {
		return result.Error
	}

	return Delete(id, user)
}
