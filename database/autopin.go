package database

import (
	"database/sql"
	"time"

	log "github.com/sirupsen/logrus"
)

func SetupAutoPinDB(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS autopin (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		channel_id TEXT NOT NULL,
		message_id TEXT NOT NULL,
		reacts INTEGER NOT NULL,
		pinned TEXT
	)`)
	if err != nil {
		log.WithField("src", "database.SetupAutoPinDB").WithError(err).Fatal("Failed to create autopin table")
	}
}

func GetAutoPin(db *sql.DB, channelId string, messageId string) (int, sql.NullString, error) {
	var reacts int
	var pinned sql.NullString
	err := db.QueryRow(`SELECT reacts, pinned FROM autopin WHERE channel_id = ? AND message_id = ?`, channelId, messageId).Scan(&reacts, &pinned)
	if err != nil {
		return 0, sql.NullString{Valid: false}, err
	}
	return reacts, pinned, nil
}

func IncrementAutoPin(db *sql.DB, channelId string, messageId string) error {

	// Check if autopin already exists
	reacts, _, err := GetAutoPin(db, channelId, messageId)
	if err == nil {
		// Update autopin
		_, err := db.Exec(`UPDATE autopin SET reacts = ? WHERE channel_id = ? AND message_id = ?`, reacts+1, channelId, messageId)
		if err != nil {
			log.WithField("src", "database.IncrementAutoPin").WithError(err).Error("Failed to update autopin")
			return err
		}

		log.WithField("src", "database.IncrementAutoPin").Info("Updated autopin")
		return nil
	}

	_, err = db.Exec(`INSERT INTO autopin (channel_id, message_id, reacts) VALUES (?, ?, ?)`, channelId, messageId, 1)
	if err != nil {
		log.WithField("src", "database.IncrementAutoPin").WithError(err).Error("Failed to add autopin")
		return err
	}

	log.WithField("src", "database.IncrementAutoPin").Info("Added autopin")
	return nil
}

func DecrementAutoPin(db *sql.DB, channelId string, messageId string) error {
	// Check if autopin already exists
	reacts, _, err := GetAutoPin(db, channelId, messageId)
	if err != nil {
		return err
	}

	newReacts := reacts - 1

	// If the new reacts is 0, delete the autopin
	if newReacts == 0 {
		_, err = db.Exec(`DELETE FROM autopin WHERE channel_id = ? AND message_id = ?`, channelId, messageId)
		if err != nil {
			log.WithField("src", "database.DecrementAutoPin").WithError(err).Error("Failed to delete autopin")
			return err
		}

		log.WithField("src", "database.DecrementAutoPin").Info("Deleted autopin")
		return nil
	}

	// Update autopin
	_, err = db.Exec(`UPDATE autopin SET reacts = ? WHERE channel_id = ? AND message_id = ?`, newReacts, channelId, messageId)
	if err != nil {
		log.WithField("src", "database.DecrementAutoPin").WithError(err).Error("Failed to update autopin")
		return err
	}

	log.WithField("src", "database.DecrementAutoPin").Info("Updated autopin")
	return nil
}

func SetAutoPinPinned(db *sql.DB, channelId string, messageId string, pinned bool) error {

	if pinned {
		_, err := db.Exec(`UPDATE autopin SET pinned = ? WHERE channel_id = ? AND message_id = ?`, time.Now().Format(time.DateTime), channelId, messageId)
		if err != nil {
			log.WithField("src", "database.SetAutoPinPinned").WithError(err).Error("Failed to update autopin")
			return err
		}
	} else {
		_, err := db.Exec(`UPDATE autopin SET pinned = NULL WHERE channel_id = ? AND message_id = ?`, channelId, messageId)
		if err != nil {
			log.WithField("src", "database.SetAutoPinPinned").WithError(err).Error("Failed to update autopin")
			return err
		}

	}

	log.WithField("src", "database.SetAutoPinPinned").Info("Updated autopin")
	return nil
}
