package autopin

import (
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SetupAutopinDB initializes the database with the Autopin model. It
// automatically migrates the database schema to match the model, ensuring the
// table is created or updated as needed.
func SetupAutopinDB(db *gorm.DB) {
	if err := db.AutoMigrate(&Autopin{}); err != nil {
		log.WithField("src", "database.SetupAutoPinDB").WithError(err).Fatal("Failed to auto-migrate autopin table")
	}
}

// GetAutopin retrieves the autopin record for a specific channel and message
// ID. It returns the number of reactions, the pinned timestamp (if any), and
// any error encountered. If the autopin is not found, gorm.ErrRecordNotFound
// is returned as the error.
func GetAutopin(db *gorm.DB, channelId, messageId string) (int, *time.Time, error) {
	var autopin Autopin
	result := db.Where("channel_id = ? AND message_id = ?", channelId, messageId).Limit(1).Find(&autopin)
	if result.Error != nil {
		return 0, nil, result.Error
	}
	return autopin.Reacts, autopin.Pinned, nil
}

// IncrementAutopin increases the reaction count for a given channel and message
// ID. If the autopin record does not exist, it creates a new one with a single
// reaction. It logs and returns any error encountered during the operation.
func IncrementAutopin(db *gorm.DB, channelId, messageId string) error {
	var autopin Autopin
	result := db.Where("channel_id = ? AND message_id = ?", channelId, messageId).FirstOrCreate(&autopin, Autopin{ChannelID: channelId, MessageID: messageId})

	if result.Error != nil {
		log.WithField("src", "database.IncrementAutoPin").WithError(result.Error).Error("Failed to find or create autopin")
		return result.Error
	}

	autopin.Reacts++
	return db.Save(&autopin).Error
}

// DecrementAutopin decreases the reaction count for a given channel and message
// ID. If the resulting reaction count is zero or less, the autopin record is
// deleted. It returns any error encountered during the find, update, or delete
// operations.
func DecrementAutopin(db *gorm.DB, channelId, messageId string) error {
	var autopin Autopin
	result := db.Where("channel_id = ? AND message_id = ?", channelId, messageId).Limit(1).Find(&autopin)
	if result.Error != nil {
		return result.Error
	}

	autopin.Reacts--
	if autopin.Reacts <= 0 {
		return db.Delete(&autopin).Error
	} else {
		return db.Save(&autopin).Error
	}
}

// SetAutopinPinned updates the pinned status of an autopin record for a given
// channel and message ID. If 'pinned' is true, it sets the current timestamp
// as the pinned time. If 'pinned' is false, it clears the pinned timestamp,
// effectively marking it as unpinned. It logs and returns any error
// encountered during the update operation.
func SetAutopinPinned(db *gorm.DB, channelId, messageId string, pinned bool) error {
	var autopin Autopin
	result := db.Where("channel_id = ? AND message_id = ?", channelId, messageId).Limit(1).Find(&autopin)
	if result.Error != nil {
		return result.Error
	}

	if pinned {
		now := time.Now()
		autopin.Pinned = &now
	} else {
		autopin.Pinned = nil
	}

	return db.Save(&autopin).Error
}
