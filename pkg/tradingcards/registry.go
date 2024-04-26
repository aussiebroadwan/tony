package tradingcards

import (
	"errors"

	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"
)

var lg *log.Entry = log.New().WithField("src", "tradingcards")

func SetupTradingCardsDB(db *gorm.DB, logger *log.Entry) {
	lg = logger

	if err := db.AutoMigrate(&UserCard{}, &Card{}); err != nil {
		lg.WithError(err).Fatal("Failed to auto-migrate tradingcards tables")
	}
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

// GetUserCard retrieves the card assigned to the user. If the card does not
// exist, it returns an error.
func GetUserCard(db *gorm.DB, userId, cardName string) (Card, error) {
	var userCards []UserCard
	err := db.Where(UserCard{UserId: userId, CardName: cardName}).Limit(1).Find(&userCards).Error
	if err != nil {
		return Card{}, err
	}

	if len(userCards) == 0 {
		return Card{}, ErrCardNotFound
	}

	// Get card information
	card, err := GetCard(db, cardName)
	if err != nil {
		return Card{}, err
	}
	card.CurrentUsage = userCards[0].Usages

	return card, err
}

// AssignCard assigns a card to a user. If the card does not exist, it returns
// an error.
func AssignCard(db *gorm.DB, userId, cardName string) error {

	// Check if card exists
	card, err := GetCard(db, cardName)
	if err != nil {
		return err
	}

	// Check if user already has the card
	var userCards []UserCard
	err = db.Where(UserCard{UserId: userId, CardName: cardName}).Limit(1).Find(&userCards).Error
	if err != nil {
		return err
	}
	if len(userCards) > 0 {
		return ErrAlreadyHaveCard
	}

	return db.Create(&UserCard{UserId: userId, CardName: cardName, Usages: card.MaxUsage}).Error
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
	var userCards []UserCard
	err = db.Where(UserCard{UserId: userId, CardName: cardName}).Limit(1).Find(&userCards).Error
	if err != nil {
		return err
	}

	if len(userCards) == 0 {
		return ErrCardNotFound
	}

	results := db.Unscoped().Delete(&(userCards[0]))
	if results.RowsAffected == 0 {
		return ErrDeleteCard
	}

	return nil
}

// TransferCard transfers a card from one user to another. If the card does not
// exist, it returns an error. If the user does not have the card, it returns an
// error.
func TransferCard(db *gorm.DB, fromUserId, toUserId, cardName string) error {
	// Check if card exists
	card, err := GetCard(db, cardName)
	if err != nil {
		return err
	}

	if !card.Tradable {
		return ErrCardNotTradable
	}

	// Check if user has the card
	_, err = GetUserCard(db, fromUserId, cardName)
	if err != nil {
		return err
	}

	// Check if to user already has the card
	_, err = GetUserCard(db, toUserId, cardName)
	if err == nil {
		return err
	}

	// Transfer the card
	fromCard := UserCard{UserId: fromUserId, CardName: cardName}
	err = db.Where(&fromCard).First(&fromCard).Error
	if err != nil {
		return err
	}

	fromCard.UserId = toUserId
	return db.Save(&fromCard).Error
}

// UseCard uses a card from the user. If the card does not exist, it returns an
// error. If the user does not have the card, it returns an error. If the card
// is unbreakable, it will not be damaged.
func UseCard(db *gorm.DB, userId, cardName string) error {
	// Check if card exists
	card, err := GetCard(db, cardName)
	if err != nil {
		return err
	}

	if !card.Usable {
		return ErrCardNotTradable
	}

	// Check if user has the card
	var ownedCards []UserCard
	err = db.Where(UserCard{UserId: userId, CardName: cardName}).Limit(1).Find(&ownedCards).Error
	if err != nil {
		return err
	}

	// Damage the card
	if !card.Unbreakable {
		ownedCards[0].Usages--
		if ownedCards[0].Usages <= 0 {
			RevokeCard(db, userId, cardName)
			return ErrCardBroken
		}
		return db.Save(&(ownedCards[0])).Error
	}

	return nil
}

// RepairCard repairs a card from the user. If the card does not exist, it returns
// an error. If the user does not have the card, it returns an error. If the card
// is unbreakable, it will not be repaired as it is already in perfect condition.
func RepairCard(db *gorm.DB, userId, cardName string, amount int) error {
	// Check if card exists
	card, err := GetCard(db, cardName)
	if err != nil {
		return err
	}

	// Check if user has the card
	var ownedCards []UserCard
	err = db.Where(UserCard{UserId: userId, CardName: cardName}).Limit(1).Find(&ownedCards).Error
	if err != nil {
		return err
	}

	if card.Unbreakable {
		return ErrCardUnbreakable
	}

	// Repair the card
	ownedCards[0].Usages += amount
	if ownedCards[0].Usages > card.MaxUsage {
		ownedCards[0].Usages = card.MaxUsage
	}

	return db.Save(&(ownedCards[0])).Error
}

// ListUserCards retrieves all cards assigned to the user.
func ListUserCards(db *gorm.DB, userId string) ([]Card, error) {
	var UserCards []UserCard
	err := db.Where(UserCard{UserId: userId}).Find(&UserCards).Error
	if err != nil {
		return nil, err
	}

	if len(UserCards) == 0 {
		return nil, ErrCardNotFound
	}

	var cards []Card
	for _, userCard := range UserCards {
		card, err := GetCard(db, userCard.CardName)
		if err != nil {
			return nil, err
		}
		card.CurrentUsage = userCard.Usages
		cards = append(cards, card)
	}

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
