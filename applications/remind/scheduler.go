package remind

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

var reminderScheduler = make(map[uint]Reminder)
var reminderStop = false

// Run periodically checks for due reminders and executes their actions
// This function should be run in a goroutine
func Run(db *gorm.DB, session *discordgo.Session) {
	ticker := time.NewTicker(10 * time.Second) // Adjust the interval as needed
	defer ticker.Stop()

	// Run the reminder loop
	for range ticker.C {
		if reminderStop {
			break
		}

		now := time.Now()
		for _, r := range reminderScheduler {
			// Check if the reminder is due
			if r.TriggerTime.Before(now) {
				r.Action(db, session)                     // Execute the reminder action
				_ = DeleteReminder(db, r.ID, r.CreatedBy) // Remove the reminder
			}
		}
	}

	// Clear the reminder store
	reminderScheduler = make(map[uint]Reminder)
	reminderStop = false
}

// Add creates a new reminder and returns its ID
func Add(id uint, r Reminder) uint {
	reminderScheduler[id] = r
	return id
}

// Delete removes a reminder by its ID
func Delete(id uint, user string) error {
	if _, ok := reminderScheduler[id]; !ok {
		return fmt.Errorf("reminder with ID %d not found", id)
	}

	if reminderScheduler[id].CreatedBy != user {
		return fmt.Errorf("reminder with ID %d does not belong to user %s", id, user)
	}

	delete(reminderScheduler, id)
	return nil
}

// List returns a slice of upcoming reminders.
func List() []Reminder {
	var upcoming []Reminder
	now := time.Now()
	for _, r := range reminderScheduler {
		if r.TriggerTime.After(now) && !r.Reminded {
			upcoming = append(upcoming, r)
		}
	}
	return upcoming
}

// Status returns the time left for a reminder.
func Status(id uint) (time.Duration, error) {
	r, ok := reminderScheduler[id]
	if !ok {
		return 0, fmt.Errorf("reminder with ID %d not found", id)
	}
	return time.Until(r.TriggerTime), nil
}
