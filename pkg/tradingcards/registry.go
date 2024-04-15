package tradingcards

import (
	"errors"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) {
	db.AutoMigrate(&UserCards{}, &Card{})
}

// RegisterCard adds a new card to the registry. If the card already exists, it
// will update the card with the new information.
func RegisterCard(db *gorm.DB, card Card) error {
	if err := card.Verify(); err != nil {
		return err
	}

	// Create card if it doesn't exist otherwise update it
	var newCard Card
	result := db.Where(Card{Name: card.Name}).Assign(card).FirstOrCreate(&newCard)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	return nil
}

// GetCard retrieves the card with the given name. If the card does not exist,
// it returns an error.
func GetCard(db *gorm.DB, cardName string) (Card, error) {
	var cards []Card
	result := db.Where(Card{Name: cardName}).Limit(1).Find(&cards)
	if result.Error != nil {
		return Card{}, ErrCardNotFound
	}

	return cards[0], nil
}

// AssignCard assigns a card to a user. If the card does not exist, it returns
// an error.
func AssignCard(db *gorm.DB, userId, cardName string) error {

	// Check if card exists
	_, err := GetCard(db, cardName)
	if err != nil {
		return err
	}

	return db.Create(&UserCards{UserId: userId, CardName: cardName}).Error
}

// GetUserCard retrieves the card assigned to the user. If the card does not
// exist, it returns an error.
func GetUserCard(db *gorm.DB, userId, cardName string) ([]Card, error) {
	var cards []Card
	err := db.Table("users_cards").
		Select("cards.*").
		Joins("join cards on cards.name = users_cards.card_name").
		Where("users_cards.user_id = ? and users_cards.card_name = ?", userId, cardName).
		Scan(&cards).Error

	return cards, err
}

// RevokeCard revokes a card from a user. If the card does not exist, it returns
// an error. If the user does not have the card, it returns an error.
func RevokeCard(db *gorm.DB, userId, cardName string) error {
	// Check if card exists
	_, err := GetCard(db, cardName)
	if err != nil {
		return err
	}

	// Check if user has the card
	_, err = GetUserCard(db, userId, cardName)
	if err != nil {
		return err
	}

	return db.Where(UserCards{UserId: userId, CardName: cardName}).Limit(1).Delete(&UserCards{}).Error
}

// ListUserCards retrieves all cards assigned to the user.
func ListUserCards(db *gorm.DB, userId string) ([]Card, error) {
	var cards []Card
	err := db.Table("users_cards").
		Select("cards.*").
		Joins("join cards on cards.name = users_cards.card_name").
		Where("users_cards.user_id = ?", userId).
		Scan(&cards).Error

	return cards, err
}

// ListApplicationCards retrieves all cards from the given application.
func ListApplicationCards(db *gorm.DB, applicationId string) ([]Card, error) {
	var cards []Card
	result := db.Where(Card{Application: applicationId}).Find(&cards)
	if result.Error != nil {
		return nil, result.Error
	}

	return cards, nil
}
