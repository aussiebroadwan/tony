package tradingcards

import "testing"

func TestValidCard(t *testing.T) {
	card := Card{
		Name:         "test_card",
		Title:        "Test Card",
		Description:  "This is a test card",
		Application:  "test",
		Rarity:       CardRarityCommon,
		Usable:       true,
		Tradable:     true,
		Unbreakable:  false,
		MaxUsage:     1,
		CurrentUsage: 0,
		SVG:          "<svg></svg>",
	}

	if err := card.Verify(); err != nil {
		t.Errorf("Expected valid card, got error: %v", err)
	}
}

func TestInvalidCard(t *testing.T) {
	card := Card{
		Name:         "test_card",
		Title:        "Test Card",
		Description:  "This is a test card",
		Application:  "test",
		Rarity:       CardRarityCommon,
		Usable:       true,
		Tradable:     true,
		Unbreakable:  false,
		MaxUsage:     0,
		CurrentUsage: 0,
		SVG:          "<svg></svg>",
	}

	if err := card.Verify(); err != ErrCardInvalidUsage {
		t.Errorf("Expected ErrCardInvalidUsage, got error: %v", err)
	}
}
