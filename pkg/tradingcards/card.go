package tradingcards

import (
	"errors"

	"gorm.io/gorm"
)

const (
	CardRarityCommon    = "common"
	CardRarityUncommon  = "uncommon"
	CardRarityRare      = "rare"
	CardRarityEpic      = "epic"
	CardRarityLegendary = "legendary"
)

var (
	ErrCardNotFound        = errors.New("card not found")
	ErrCardExists          = errors.New("card already exists")
	ErrCardNameRequired    = errors.New("card name is required")
	ErrCardNameTooLong     = errors.New("card name is too long")
	ErrCardInfoRequired    = errors.New("card title, description is required")
	ErrCardInfoTooLong     = errors.New("card title, description is too long")
	ErrCardInvalidUsage    = errors.New("invalid has card usage")
	ErrCardRarityInvalid   = errors.New("invalid card rarity")
	ErrApplicationRequired = errors.New("card application is required")
	ErrApplicationTooLong  = errors.New("card application is too long")
	ErrCardNotTradable     = errors.New("card is not tradable")
	ErrAlreadyHaveCard     = errors.New("user already has card")
	ErrDeleteCard          = errors.New("cant delete card from user")
	ErrCardUnbreakable     = errors.New("card is unbreakable")
)

type UserCard struct {
	gorm.Model

	UserId   string
	CardName string

	Usages int
}

type Card struct {
	Name string // Unique identifier ie. `blackjack_achievement_big_win`

	Title       string
	Description string `gorm:"size:1024"`
	Application string // Application ID where did it come from, ie. `blackjack`

	Rarity string

	// Flags
	Usable      bool
	Tradable    bool
	Unbreakable bool

	// Usage
	MaxUsage     int
	CurrentUsage int `gorm:"-"` // Not stored in database filled in API

	// Graphic
	SVG string
}

func (c Card) Verify() error {
	if c.Name == "" {
		return ErrCardNameRequired
	}

	if len(c.Name) > 255 {
		return ErrCardNameTooLong
	}

	if c.Application == "" {
		return ErrApplicationRequired
	}

	if len(c.Application) > 255 {
		return ErrApplicationTooLong
	}

	if c.Title == "" || c.Description == "" {
		return ErrCardInfoRequired
	}

	if len(c.Title) > 255 || len(c.Description) > 1024 {
		return ErrCardInfoTooLong
	}

	if c.Rarity != CardRarityCommon && c.Rarity != CardRarityUncommon && c.Rarity != CardRarityRare && c.Rarity != CardRarityEpic && c.Rarity != CardRarityLegendary {
		return ErrCardRarityInvalid
	}

	if c.Usable && c.MaxUsage <= 0 && !c.Unbreakable {
		return ErrCardInvalidUsage
	}

	return nil
}
