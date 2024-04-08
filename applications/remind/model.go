package remind

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Reminder struct {
	gorm.Model  // Includes fields ID, CreatedAt, UpdatedAt, DeletedAt
	CreatedBy   string
	ChannelID   string
	TriggerTime time.Time
	Message     string
	Reminded    bool
}

func (r Reminder) Action(db *gorm.DB, session *discordgo.Session) {
	// Send the reminder message
	session.ChannelMessageSend(r.ChannelID, fmt.Sprintf("%s %s", r.CreatedBy, r.Message))

	// Set the reminder as reminded
	r.Reminded = true
	if err := db.Save(&r).Error; err != nil {
		log.Errorf("Failed to mark reminder %d as reminded: %v", r.ID, err)
	}
}
