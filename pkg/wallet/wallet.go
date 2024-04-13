package wallet

import (
	"errors"
	"sync"

	lg "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	// DefaultBalance is the default balance for a new user.
	DefaultBalance = 500
)

var (
	mu sync.Mutex
)

type User struct {
	gorm.Model

	UserId  string `gorm:"unique"` // Discord User ID
	Balance int64
}

type TransactionType string

const (
	CREDIT TransactionType = "CREDIT"
	DEBIT  TransactionType = "DEBIT"
)

type Transaction struct {
	gorm.Model

	// Transaction Type
	Type   TransactionType `gorm:"type:string;not null"`
	Amount int64

	// Transaction Metadata
	Description   string
	ApplicationId string

	// Owner of the wallet that the transaction is related to
	UserId uint
	User   User `gorm:"foreignKey:UserId"`
}

// SetupWalletDB initializes the database with the User and Transaction models. It
// automatically migrates the database schema to match the models, ensuring the
// tables are created or updated as needed.
func SetupWalletDB(db *gorm.DB) {
	mu = sync.Mutex{}

	if err := db.AutoMigrate(&User{}, &Transaction{}); err != nil {
		lg.WithField("src", "database.SetupWalletDB").WithError(err).Fatal("Failed to auto-migrate wallet tables")
	}
}

// getUser retrieves the user with the given ID. If the user does not exist, it
// creates a new user with the default balance and returns a user with the
// default balance.
func getUser(db *gorm.DB, userId string) (User, error) {
	var user User
	result := db.Where(User{UserId: userId}).Attrs(User{Balance: DefaultBalance}).FirstOrCreate(&user)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return user, result.Error
	}

	return user, nil
}

// createTransaction creates a new transaction with the given type, amount,
// description, and application ID. It logs and returns any error encountered
// during the operation.
func createTransaction(db *gorm.DB, transactionType TransactionType, amount int64, description, applicationId string, userId uint) error {
	transaction := Transaction{
		Type:          transactionType,
		Amount:        amount,
		Description:   description,
		ApplicationId: applicationId,
		UserId:        userId,
	}

	result := db.Create(&transaction)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// Balance retrieves the balance of a user with the given ID. If the user is
// not found, initialise a new user with the default balance and return the
// default balance.
func Balance(db *gorm.DB, userId string) (int64, error) {
	mu.Lock()
	defer mu.Unlock()

	user, err := getUser(db, userId)
	if err != nil {
		return 0, err
	}

	return user.Balance, nil
}

// Credit adds the specified amount to the balance of the user with the given
// ID. It logs and returns any error encountered during the operation. If the
// user does not exist, it creates a new user with the default balance and
// credits the specified amount.
func Credit(db *gorm.DB, userId string, amount int64, description, applicationId string) error {
	mu.Lock()
	defer mu.Unlock()

	user, err := getUser(db, userId)
	if err != nil {
		return err
	}

	user.Balance += amount
	if err := db.Save(&user).Error; err != nil {
		return err
	}

	return createTransaction(db, CREDIT, amount, description, applicationId, user.ID)
}

// Debit subtracts the specified amount from the balance of the user with the
// given ID. It logs and returns any error encountered during the operation. If
// the user does not exist, it creates a new user with the default balance and
// debits the specified amount.
func Debit(db *gorm.DB, userId string, amount int64, description, applicationId string) error {
	mu.Lock()
	defer mu.Unlock()

	user, err := getUser(db, userId)
	if err != nil {
		return err
	}

	if user.Balance < amount {
		return errors.New("insufficient balance")
	}

	user.Balance -= amount
	if err := db.Save(&user).Error; err != nil {
		return err
	}

	return createTransaction(db, DEBIT, amount, description, applicationId, user.ID)
}

// History retrieves the transaction history of the user with the given ID. It
// returns the last 'limit' number of transactions. If 'limit' is negative, it
// returns all transactions.
func History(db *gorm.DB, userId string, limit int) ([]Transaction, error) {
	mu.Lock()
	defer mu.Unlock()

	// Default limit to 10 if not provided
	if limit == 0 {
		limit = 10
	}

	// If limit is negative, return all transactions
	if limit <= 0 {
		limit = -1
	}

	var transactions []Transaction
	result := db.Where(Transaction{User: User{UserId: userId}}).Limit(limit).Find(&transactions)
	if result.Error != nil {
		return nil, result.Error
	}

	return transactions, nil
}

// CreditHistory retrieves the credit transaction history of the user with the
// given ID. It returns the last 'limit' number of credit transactions. If
// 'limit' is negative, it returns all credit transactions.
func CreditHistory(db *gorm.DB, userId string, limit int) ([]Transaction, error) {
	mu.Lock()
	defer mu.Unlock()

	// Default limit to 10 if not provided
	if limit == 0 {
		limit = 10
	}

	// If limit is negative, return all transactions
	if limit <= 0 {
		limit = -1
	}

	var transactions []Transaction
	result := db.Where(Transaction{Type: CREDIT, User: User{UserId: userId}}).Limit(limit).Find(&transactions)
	if result.Error != nil {
		return nil, result.Error
	}

	return transactions, nil
}

// DebitHistory retrieves the debit transaction history of the user with the
// given ID. It returns the last 'limit' number of debit transactions. If
// 'limit' is negative, it returns all debit transactions.
func DebitHistory(db *gorm.DB, userId string, limit int) ([]Transaction, error) {
	mu.Lock()
	defer mu.Unlock()

	// Default limit to 10 if not provided
	if limit == 0 {
		limit = 10
	}

	// If limit is negative, return all transactions
	if limit <= 0 {
		limit = -1
	}

	var transactions []Transaction
	result := db.Where(Transaction{Type: DEBIT, User: User{UserId: userId}}).Limit(limit).Find(&transactions)
	if result.Error != nil {
		return nil, result.Error
	}

	return transactions, nil
}
