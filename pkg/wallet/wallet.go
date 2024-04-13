package wallet

import (
	"errors"
	"fmt"
	"sync"

	lg "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var mu sync.Mutex

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
)

// SetupWalletDB initializes the database with the User and Transaction models. It
// automatically migrates the database schema to match the models, ensuring the
// tables are created or updated as needed.
func SetupWalletDB(db *gorm.DB) {
	mu = sync.Mutex{}

	if err := db.AutoMigrate(&Transaction{}, &WalletUser{}); err != nil {
		lg.WithField("src", "database.SetupWalletDB").WithError(err).Fatal("Failed to auto-migrate wallet tables")
	}
}

// getUser retrieves the user with the given ID. If the user does not exist, it
// creates a new user with the default balance and returns a user with the
// default balance.
func getUser(db *gorm.DB, userId string) (WalletUser, error) {
	var user WalletUser
	result := db.Where(WalletUser{UserId: userId}).Attrs(WalletUser{Balance: DefaultBalance}).FirstOrCreate(&user)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return user, result.Error
	}

	return user, nil
}

// createTransaction creates a new transaction with the given type, amount,
// description, and application ID. It logs and returns any error encountered
// during the operation.
func createTransaction(db *gorm.DB, transactionType TransactionType, amount int64, description, applicationId string, userId string) error {
	transaction := Transaction{
		Type:          transactionType,
		Amount:        amount,
		Description:   description,
		ApplicationId: applicationId,
		UserID:        userId,
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

	return createTransaction(db, CREDIT, amount, description, applicationId, user.UserId)
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
		return ErrInsufficientBalance
	}

	user.Balance -= amount
	if err := db.Save(&user).Error; err != nil {
		return err
	}

	return createTransaction(db, DEBIT, amount, description, applicationId, user.UserId)
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
	result := db.Where(Transaction{UserID: userId}).Order("created_at desc").Limit(limit).Find(&transactions)
	if result.Error != nil {
		return nil, result.Error
	}

	for i := range transactions {
		fmt.Printf("Transaction: %+v\n", transactions[i])
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
	result := db.Where(Transaction{Type: CREDIT, UserID: userId}).Order("created_at desc").Limit(limit).Find(&transactions)
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
	result := db.Where(Transaction{Type: DEBIT, UserID: userId}).Order("created_at desc").Limit(limit).Find(&transactions)
	if result.Error != nil {
		return nil, result.Error
	}

	return transactions, nil
}
