package remind

import (
	"fmt"
	"time"
)

type Reminder struct {
	ID int64

	CreatedBy string
	Triggered bool

	TriggerTime time.Time
	Action      func()
}

var reminderStore = make(map[int64]Reminder)
var reminderStop = false

// Load initialises the reminder store with the provided map, this is useful
// for testing and also for loading reminders from a database
func Load(store map[int64]Reminder) {
	reminderStore = store
}

// Run periodically checks for due reminders and executes their actions
// This function should be run in a goroutine
func Run() {
	ticker := time.NewTicker(10 * time.Second) // Adjust the interval as needed
	defer ticker.Stop()

	// Run the reminder loop
	for range ticker.C {
		if reminderStop {
			break
		}

		now := time.Now()
		for _, r := range reminderStore {
			// Check if the reminder is due
			if r.TriggerTime.Before(now) {
				r.Action()                    // Execute the reminder action
				_ = Delete(r.ID, r.CreatedBy) // Remove the reminder
			}
		}
	}

	// Clear the reminder store
	reminderStore = make(map[int64]Reminder)
	reminderStop = false
}

func Stop() {
	reminderStop = true
}

// Add creates a new reminder and returns its ID
func Add(id int64, triggerTime time.Time, createdBy string, action func()) int64 {
	reminderStore[id] = Reminder{
		ID:          id,
		CreatedBy:   createdBy,
		TriggerTime: triggerTime,
		Triggered:   false,
		Action:      action,
	}

	return id
}

// Delete removes a reminder by its ID
func Delete(id int64, user string) error {
	if _, ok := reminderStore[id]; !ok {
		return fmt.Errorf("reminder with ID %d not found", id)
	}

	if reminderStore[id].CreatedBy != user {
		return fmt.Errorf("reminder with ID %d does not belong to user %s", id, user)
	}

	delete(reminderStore, id)
	return nil
}

// List returns a slice of upcoming reminders.
func List() []Reminder {
	var upcoming []Reminder
	now := time.Now()
	for _, r := range reminderStore {
		if r.TriggerTime.After(now) && !r.Triggered {
			upcoming = append(upcoming, r)
		}
	}
	return upcoming
}

// Status returns the time left for a reminder.
func Status(id int64) (time.Duration, error) {
	r, ok := reminderStore[id]
	if !ok {
		return 0, fmt.Errorf("reminder with ID %d not found", id)
	}
	return time.Until(r.TriggerTime), nil
}
