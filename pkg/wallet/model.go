package wallet

import "gorm.io/gorm"

// DefaultBalance is the default balance for a new user.
const DefaultBalance = 500

type TransactionType string

const (
	CREDIT TransactionType = "CREDIT"
	DEBIT  TransactionType = "DEBIT"
)

type WalletUser struct {
	UserId  string `gorm:"primarykey"` // Discord User ID
	Balance int64
}

type Transaction struct {
	gorm.Model

	// Transaction Type
	Type   TransactionType `gorm:"type:string;not null"`
	Amount int64

	// Transaction Metadata
	Description   string
	ApplicationId string

	// Owner of the wallet that the transaction is related to
	UserID string
	User   WalletUser `gorm:"references:UserId"`
}
