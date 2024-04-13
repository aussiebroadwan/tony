package wallet

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const ExampleUserId1 = "1060681976622891089"
const ExampleUserId2 = "169015299834642432"

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&User{}, &Transaction{}); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	return db
}

func TestGetUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Migrator().DropTable(&User{})

	// Test case: Retrieve an existing user
	db.Create(&User{UserId: ExampleUserId1, Balance: DefaultBalance})

	user, err := getUser(db, ExampleUserId1)
	if err != nil || user.UserId != ExampleUserId1 || user.Balance != DefaultBalance {
		t.Errorf("getUser failed to retrieve existing user: %v", err)
	}

	// Test case: Create a new user if not exists
	newUser, err := getUser(db, ExampleUserId2)
	if err != nil || newUser.UserId != ExampleUserId2 || newUser.Balance != DefaultBalance {
		t.Errorf("getUser failed to create new user: %v", err)
	}
}

func TestCredit(t *testing.T) {
	db := setupTestDB(t)
	defer db.Migrator().DropTable(&User{}, &Transaction{})

	user := User{UserId: ExampleUserId1, Balance: DefaultBalance}
	db.Create(&user)

	// Test adding credit
	err := Credit(db, ExampleUserId1, 100, "test credit", "app1")
	if err != nil {
		t.Errorf("Credit failed: %v", err)
	}

	// Check balance update
	balance, err := Balance(db, ExampleUserId1)
	if err != nil {
		t.Errorf("Balance failed: %v", err)
	}

	if balance != DefaultBalance+100 {
		t.Errorf("Debit did not update balance correctly: expected %d, got %d", DefaultBalance+100, balance)
	}

	// Test transaction record
	tx := Transaction{}
	if err := db.First(&tx, Transaction{User: User{UserId: ExampleUserId1}}).Error; err != nil {
		t.Errorf("Transaction not recorded: %v", err)
	}

	if tx.Type != CREDIT || tx.Amount != 100 {
		t.Errorf("Incorrect transaction details: expected %v, got %v", CREDIT, tx.Type)
	}
}

func TestCreditNoUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Migrator().DropTable(&User{}, &Transaction{})

	// Test case: Create a new user if not exists
	err := Credit(db, ExampleUserId2, 100, "test credit", "app1")
	if err != nil {
		t.Errorf("Credit failed: %v", err)
	}

	// Check balance update
	balance, err := Balance(db, ExampleUserId2)
	if err != nil {
		t.Errorf("Balance failed: %v", err)
	}

	if balance != DefaultBalance+100 {
		t.Errorf("Debit did not update balance correctly: expected %d, got %d", DefaultBalance+100, balance)
	}
}

func TestDebit(t *testing.T) {
	db := setupTestDB(t)
	defer db.Migrator().DropTable(&User{}, &Transaction{})

	user := User{UserId: ExampleUserId1, Balance: DefaultBalance}
	db.Create(&user)

	// Test adding credit
	err := Debit(db, ExampleUserId1, 100, "test debit", "app1")
	if err != nil {
		t.Errorf("Debit failed: %v", err)
	}

	// Check balance update
	balance, err := Balance(db, ExampleUserId1)
	if err != nil {
		t.Errorf("Balance failed: %v", err)
	}

	if balance != DefaultBalance-100 {
		t.Errorf("Debit did not update balance correctly: expected %d, got %d", DefaultBalance-100, balance)
	}

	// Test transaction record
	tx := Transaction{}
	if err := db.First(&tx, Transaction{User: User{UserId: ExampleUserId1}}).Error; err != nil {
		t.Errorf("Transaction not recorded: %v", err)
	}

	if tx.Type != DEBIT || tx.Amount != 100 {
		t.Errorf("Incorrect transaction details: expected %v, got %v", DEBIT, tx.Type)
	}
}

func TestDebitNoUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Migrator().DropTable(&User{}, &Transaction{})

	// Test case: Create a new user if not exists
	err := Debit(db, ExampleUserId2, 100, "test debit", "app1")
	if err != nil {
		t.Errorf("Debit failed: %v", err)
	}

	// Check balance update
	balance, err := Balance(db, ExampleUserId2)
	if err != nil {
		t.Errorf("Balance failed: %v", err)
	}

	if balance != DefaultBalance-100 {
		t.Errorf("Debit did not update balance correctly: expected %d, got %d", DefaultBalance-100, balance)
	}
}

func setupTestDBWithTransactions(t *testing.T) *gorm.DB {
	db := setupTestDB(t) // Assuming setupTestDB is the same function provided in the previous response.

	// Creating test user
	user := User{UserId: "user123", Balance: 1000}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Creating test transactions
	transactions := []Transaction{
		{Type: CREDIT, Amount: 300, UserId: user.ID},
		{Type: CREDIT, Amount: 200, UserId: user.ID},
		{Type: DEBIT, Amount: 150, UserId: user.ID},
		{Type: DEBIT, Amount: 50, UserId: user.ID},
	}
	for _, tx := range transactions {
		if err := db.Create(&tx).Error; err != nil {
			t.Fatalf("failed to create test transaction: %v", err)
		}
	}

	return db
}

func TestHistory(t *testing.T) {
	db := setupTestDBWithTransactions(t)
	defer db.Migrator().DropTable(&User{}, &Transaction{})

	// Test fetching limited transactions
	transactions, err := History(db, "user123", 2)
	if err != nil || len(transactions) != 2 {
		t.Errorf("Expected 2 transactions, got %d, error: %v", len(transactions), err)
	}

	// Test fetching all transactions with negative limit
	transactions, err = History(db, "user123", -1)
	if err != nil || len(transactions) != 4 {
		t.Errorf("Expected 4 transactions, got %d, error: %v", len(transactions), err)
	}
}

func TestCreditHistory(t *testing.T) {
	db := setupTestDBWithTransactions(t)
	defer db.Migrator().DropTable(&User{}, &Transaction{})

	// Test fetching credit transactions
	transactions, err := CreditHistory(db, "user123", -1)
	if err != nil || len(transactions) != 2 {
		t.Errorf("Expected 2 credit transactions, got %d, error: %v", len(transactions), err)
	}
}

func TestDebitHistory(t *testing.T) {
	db := setupTestDBWithTransactions(t)
	defer db.Migrator().DropTable(&User{}, &Transaction{})

	// Test fetching debit transactions
	transactions, err := DebitHistory(db, "user123", -1)
	if err != nil || len(transactions) != 2 {
		t.Errorf("Expected 2 debit transactions, got %d, error: %v", len(transactions), err)
	}
}
