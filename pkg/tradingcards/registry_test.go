package tradingcards

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	return db
}

func TestRegisterCard(t *testing.T) {
	db := setupTestDB(t)
	defer db.Migrator().DropTable(&UserCard{}, &Card{})

	card := Card{
		Name:        "test_card",
		Title:       "Test Card",
		Description: "This is a test card",
		Application: "test",
		Rarity:      CardRarityCommon,
		Usable:      true,
		Tradable:    true,
		Unbreakable: false,
		MaxUsage:    1,
		SVG:         "<svg></svg>",
	}

	if err := RegisterCard(db, card); err != nil {
		t.Errorf("RegisterCard failed: %v", err)
		return
	}

	// Get Card
	_, err := GetCard(db, card.Name)
	if err != nil {
		t.Errorf("GetCard failed: %v", err)
	}
}

func TestCardAssign(t *testing.T) {
	db := setupTestDB(t)
	defer db.Migrator().DropTable(&UserCard{}, &Card{})

	card := Card{
		Name:        "test_card",
		Title:       "Test Card",
		Description: "This is a test card",
		Application: "test",
		Rarity:      CardRarityCommon,
		Usable:      true,
		Tradable:    true,
		Unbreakable: false,
		MaxUsage:    1,
		SVG:         "<svg></svg>",
	}

	if err := RegisterCard(db, card); err != nil {
		t.Errorf("RegisterCard failed: %v", err)
		return
	}

	// Assign Card
	err := AssignCard(db, "1", card.Name)
	if err != nil {
		t.Errorf("AssignCard failed: %v", err)
		return
	}

	// Get User Card
	_, err = GetUserCard(db, "1", card.Name)
	if err != nil {
		t.Errorf("GetUserCard failed: %v", err)
	}
}

func TestCardRevoke(t *testing.T) {
	db := setupTestDB(t)
	defer db.Migrator().DropTable(&UserCard{}, &Card{})

	card := Card{
		Name:        "test_card",
		Title:       "Test Card",
		Description: "This is a test card",
		Application: "test",
		Rarity:      CardRarityCommon,
		Usable:      true,
		Tradable:    true,
		Unbreakable: false,
		MaxUsage:    1,
		SVG:         "<svg></svg>",
	}

	if err := RegisterCard(db, card); err != nil {
		t.Errorf("RegisterCard failed: %v", err)
		return
	}

	// Assign Card
	err := AssignCard(db, "1", card.Name)
	if err != nil {
		t.Errorf("AssignCard failed: %v", err)
		return
	}

	// Revoke Card
	err = RevokeCard(db, "1", card.Name)
	if err != nil {
		t.Errorf("RevokeCard failed: %v", err)
		return
	}

	// Get User Card
	userCard, err := GetUserCard(db, "1", card.Name)
	if err == nil {
		t.Errorf("GetUserCards failed: %v with %+v", err, userCard)
	}
}

func TestTransferCard(t *testing.T) {
	db := setupTestDB(t)
	defer db.Migrator().DropTable(&UserCard{}, &Card{})

	card := Card{
		Name:        "test_card",
		Title:       "Test Card",
		Description: "This is a test card",
		Application: "test",
		Rarity:      CardRarityCommon,
		Usable:      true,
		Tradable:    true,
		Unbreakable: false,
		MaxUsage:    1,
		SVG:         "<svg></svg>",
	}

	if err := RegisterCard(db, card); err != nil {
		t.Errorf("RegisterCard failed: %v", err)
		return
	}

	// Assign Card
	err := AssignCard(db, "1", card.Name)
	if err != nil {
		t.Errorf("AssignCard failed: %v", err)
		return
	}

	// Transfer Card
	err = TransferCard(db, "1", "2", card.Name)
	if err != nil {
		t.Errorf("TransferCard failed: %v", err)
		return
	}

	// Get User Card
	fromCard, err := GetUserCard(db, "1", card.Name)
	if err == nil {
		t.Errorf("GetUserCards failed: %v with %+v", err, fromCard)
		return
	}

	// Get User Card
	toCard, err := GetUserCard(db, "2", card.Name)
	if err != nil {
		t.Errorf("GetUserCards failed: %v with %+v", err, toCard)
	}
}

func TestListUserCards(t *testing.T) {
	db := setupTestDB(t)
	defer db.Migrator().DropTable(&UserCard{}, &Card{})

	card := Card{
		Name:        "test_card",
		Title:       "Test Card",
		Description: "This is a test card",
		Application: "test",
		Rarity:      CardRarityCommon,
		Usable:      true,
		Tradable:    true,
		Unbreakable: false,
		MaxUsage:    1,
		SVG:         "<svg></svg>",
	}

	if err := RegisterCard(db, card); err != nil {
		t.Errorf("RegisterCard failed: %v", err)
		return
	}

	// Assign Card
	err := AssignCard(db, "1", card.Name)
	if err != nil {
		t.Errorf("AssignCard failed: %v", err)
		return
	}

	// List User Cards
	cards, err := ListUserCards(db, "1")
	if err != nil {
		t.Errorf("ListUserCards failed: %v", err)
		return
	}

	if len(cards) != 1 {
		t.Errorf("Expected 1 card, got %d", len(cards))
		return
	}

	if cards[0].Name != card.Name {
		t.Errorf("Expected card name %s, got %s", card.Name, cards[0].Name)
	}
}

func TestUseCard(t *testing.T) {
	db := setupTestDB(t)
	defer db.Migrator().DropTable(&UserCard{}, &Card{})

	card := Card{
		Name:        "test_card",
		Title:       "Test Card",
		Description: "This is a test card",
		Application: "test",
		Rarity:      CardRarityCommon,
		Usable:      true,
		Tradable:    true,
		Unbreakable: false,
		MaxUsage:    2,
		SVG:         "<svg></svg>",
	}

	if err := RegisterCard(db, card); err != nil {
		t.Errorf("RegisterCard failed: %v", err)
		return
	}

	// Assign Card
	err := AssignCard(db, "1", card.Name)
	if err != nil {
		t.Errorf("AssignCard failed: %v", err)
		return
	}

	// Use Card
	err = UseCard(db, "1", card.Name)
	if err != nil {
		t.Errorf("UseCard failed: %v", err)
		return
	}

	// Get User Card
	userCard, err := GetUserCard(db, "1", card.Name)
	if err != nil {
		t.Errorf("GetUserCards failed: %v", err)
		return
	}

	if userCard.CurrentUsage != 1 {
		t.Errorf("Expected current usage 1, got %d", userCard.CurrentUsage)
		return
	}

	// Use Card
	err = UseCard(db, "1", card.Name)
	if err != nil {
		t.Errorf("UseCard failed: %v", err)
		return
	}

	// Get User Card
	userCard, err = GetUserCard(db, "1", card.Name)
	if err != ErrCardNotFound {
		t.Errorf("GetUserCards failed: %v", err)
		return
	}
}

func TestRepairCard(t *testing.T) {
	db := setupTestDB(t)
	defer db.Migrator().DropTable(&UserCard{}, &Card{})

	card := Card{
		Name:        "test_card",
		Title:       "Test Card",
		Description: "This is a test card",
		Application: "test",
		Rarity:      CardRarityCommon,
		Usable:      true,
		Tradable:    true,
		Unbreakable: false,
		MaxUsage:    2,
		SVG:         "<svg></svg>",
	}

	if err := RegisterCard(db, card); err != nil {
		t.Errorf("RegisterCard failed: %v", err)
		return
	}

	// Assign Card
	err := AssignCard(db, "1", card.Name)
	if err != nil {
		t.Errorf("AssignCard failed: %v", err)
		return
	}

	// Use Card
	err = UseCard(db, "1", card.Name)
	if err != nil {
		t.Errorf("UseCard failed: %v", err)
		return
	}

	// Repair Card
	err = RepairCard(db, "1", card.Name, 5)
	if err != nil {
		t.Errorf("RepairCard failed: %v", err)
		return
	}

	// Get User Card
	userCard, err := GetUserCard(db, "1", card.Name)
	if err != nil {
		t.Errorf("GetUserCards failed: %v", err)
		return
	}

	if userCard.CurrentUsage != card.MaxUsage {
		t.Errorf("Expected current usage 0, got %d", userCard.CurrentUsage)
		return
	}
}
