package remind

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"

	log "github.com/sirupsen/logrus"

	_ "github.com/mattn/go-sqlite3"
)

func SetupRemindersDB(db *sql.DB, session *discordgo.Session) {
	// Create the reminders table
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS reminders (
		id INTEGER PRIMARY KEY,
		created_by TEXT NOT NULL,
		channel_id TEXT NOT NULL,
		trigger_time TEXT NOT NULL,
		message TEXT NOT NULL,
		reminded BOOLEAN DEFAULT FALSE
	)`)
	if err != nil {
		log.WithField("src", "database.SetupRemindersDB").WithError(err).Fatal("Failed to create reminders table")
	}

	// Load all reminders from the database
	rows, err := db.Query(`SELECT id, created_by, channel_id, trigger_time, message, reminded FROM reminders`)
	if err != nil {
		log.WithField("src", "database.SetupRemindersDB").WithError(err).Fatal("Failed to load reminders from database")
	}

	var loadReminders = make(map[int64]Reminder)

	// Iterate over each reminder
	for rows.Next() {
		var id int64
		var createdBy, channelId, triggerTime, message string
		var reminded bool

		err := rows.Scan(&id, &createdBy, &channelId, &triggerTime, &message, &reminded)
		if err != nil {
			log.WithField("src", "database.SetupRemindersDB").WithError(err).Error("Failed to scan reminder row")
			continue
		}

		// Parse the trigger time
		t, err := time.ParseInLocation(time.DateTime, triggerTime, time.Local)
		if err != nil {
			log.WithField("src", "database.SetupRemindersDB").WithError(err).Error("Failed to parse trigger time")
			continue
		}

		// If the reminder has already been reminded, skip it
		if reminded {
			loadReminders[id] = Reminder{
				ID:          id,
				CreatedBy:   createdBy,
				TriggerTime: t,

				Triggered: true,
				Action:    func() { /* Do nothing */ },
			}
			continue
		}

		// Add the reminder to the reminders package
		loadReminders[id] = Reminder{
			ID:          id,
			CreatedBy:   createdBy,
			TriggerTime: t,

			Triggered: false,
			Action: func() {
				// Send the reminder message
				session.ChannelMessageSend(channelId, fmt.Sprintf("%s %s", createdBy, message))

				// Set the reminder as reminded
				_, err := db.Exec(`UPDATE reminders SET reminded = TRUE WHERE id = ?`, id)
				if err != nil {
					log.WithField("src", "database.SetupRemindersDB").WithError(err).Errorf("Failed to mark reminder %d as reminded", id)
				}
			},
		}
	}

	// Load the reminders into the reminders package
	Load(loadReminders)
}

func AddReminder(db *sql.DB, createdBy string, triggerTime time.Time, session *discordgo.Session, channelId string, message string) (int64, error) {
	query := fmt.Sprintf(`INSERT INTO reminders (created_by, channel_id, trigger_time, message) VALUES ("%s", "%s", "%s", "%s")`, createdBy, channelId, triggerTime.Format(time.DateTime), message)
	result, err := db.Exec(query)

	if err != nil {
		return 0, err
	}

	// Get the ID of the reminder
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Add the reminder to the reminders package
	Add(id, triggerTime, createdBy, func() {
		// Send the reminder message
		session.ChannelMessageSend(channelId, fmt.Sprintf("%s %s", createdBy, message))

		// Set the reminder as reminded
		_, err := db.Exec(`UPDATE reminders SET reminded = TRUE WHERE id = ?`, id)
		if err != nil {
			log.WithField("src", "database.AddReminder").WithError(err).Errorf("Failed to mark reminder %d as reminded", id)
		}
	})

	return id, err
}

func DeleteReminder(db *sql.DB, id int64, user string) error {
	err := Delete(id, user)
	if err != nil {
		return err
	}

	_, err = db.Exec(`UPDATE reminders SET reminded = TRUE WHERE id = ?`, id)
	return err
}
