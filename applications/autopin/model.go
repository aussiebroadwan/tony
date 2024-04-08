package autopin

import (
	"time"

	"gorm.io/gorm"
)

type Autopin struct {
	gorm.Model // Includes fields like ID, CreatedAt, UpdatedAt, which you may or may not want to use.
	ChannelID  string
	MessageID  string
	Reacts     int
	Pinned     *time.Time // Use *time.Time to allow for a nullable timestamp
}
