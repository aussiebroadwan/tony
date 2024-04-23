package blackjack_app

import (
	"github.com/aussiebroadwan/tony/pkg/blackjack"
	"github.com/aussiebroadwan/tony/pkg/tradingcards"
	"gorm.io/gorm"
)

const (
	applicationId = "blackjack"
)

func RegisterCards(db *gorm.DB) {
	tradingcards.RegisterCard(db, FirstTimeWinnerCard)
	tradingcards.RegisterCard(db, VeteranPlayerCard)
	tradingcards.RegisterCard(db, BlackjackStreakCard)
	tradingcards.RegisterCard(db, HighRollerCard)
	tradingcards.RegisterCard(db, CombackKingCard)
	tradingcards.RegisterCard(db, Perfect21Card)
	tradingcards.RegisterCard(db, LuckySevenCard)
}

var Cards = map[string]tradingcards.Card{
	blackjack.FirstTimeWinner: FirstTimeWinnerCard,
	blackjack.VeteranPlayer:   VeteranPlayerCard,
	blackjack.BlackjackStreak: BlackjackStreakCard,
	blackjack.HighRoller:      HighRollerCard,
	blackjack.OhShit:          OhShitCard,
	blackjack.CombackKing:     CombackKingCard,
	blackjack.Perfect21:       Perfect21Card,
	blackjack.LuckySeven:      LuckySevenCard,
}

var FirstTimeWinnerCard = tradingcards.Card{
	Name:        blackjack.FirstTimeWinner,
	Application: applicationId,
	Title:       "First Time Winner",
	Description: "Awarded for winning your first game at the Blackjack tables, this card marks the beginning of your gambling journey.",
	Rarity:      tradingcards.CardRarityCommon,
	Tradable:    false,
	Usable:      false,
	Unbreakable: true,
	SVG:         "",
}

var VeteranPlayerCard = tradingcards.Card{
	Name:        blackjack.VeteranPlayer,
	Application: applicationId,
	Title:       "Veteran Player",
	Description: "Issued to those who have played over 100 shoes, this card honors your dedication and long-standing participation.",
	Rarity:      tradingcards.CardRarityUncommon,
	Tradable:    false,
	Usable:      false,
	Unbreakable: true,
	SVG:         "",
}

var BlackjackStreakCard = tradingcards.Card{
	Name:        blackjack.BlackjackStreak,
	Application: applicationId,
	Title:       "Blackjack Streak",
	Description: "Earn this card by hitting three consecutive blackjacks in one shoe, a testament to your skill and good fortune.",
	Rarity:      tradingcards.CardRarityRare,
	Tradable:    false,
	Usable:      false,
	Unbreakable: true,
	SVG:         "",
}

var HighRollerCard = tradingcards.Card{
	Name:        blackjack.HighRoller,
	Application: applicationId,
	Title:       "High Roller",
	Description: "This card celebrates your achievement of accumulating over 1,000 credits in profit, distinguishing you as one of the elite players.",
	Rarity:      tradingcards.CardRarityRare,
	Tradable:    false,
	Usable:      false,
	Unbreakable: true,
	SVG:         "",
}

var OhShitCard = tradingcards.Card{
	Name:        blackjack.OhShit,
	Application: applicationId,
	Title:       "Oh Shit!",
	Description: "Congratulations, you've just pissed away 1,000 credits.",
	Rarity:      tradingcards.CardRarityEpic,
	Tradable:    false,
	Usable:      false,
	Unbreakable: true,
	SVG:         "",
}

var CombackKingCard = tradingcards.Card{
	Name:        blackjack.CombackKing,
	Application: applicationId,
	Title:       "Comeback King",
	Description: "This card is awarded to players who turn a game around from a long losing streak to win with a blackjack. Celebrate your resilience and strategic comeback.",
	Rarity:      tradingcards.CardRarityEpic,
	Tradable:    false,
	Usable:      false,
	Unbreakable: true,
	SVG:         "",
}

var Perfect21Card = tradingcards.Card{
	Name:        blackjack.Perfect21,
	Application: applicationId,
	Title:       "Perfect 21",
	Description: "A legendary card for those skilled enough to achieve a perfect 21 total 21 times, showcasing your expert level of luck.",
	Rarity:      tradingcards.CardRarityLegendary,
	Tradable:    false,
	Usable:      false,
	Unbreakable: true,
	SVG:         "",
}

var LuckySevenCard = tradingcards.Card{
	Name:        blackjack.LuckySeven,
	Application: applicationId,
	Title:       "Lucky Seven",
	Description: "Granted to players who win a game with a hand totaling exactly 21, using seven cards. A rare display of patience and luck.",
	Rarity:      tradingcards.CardRarityLegendary,
	Tradable:    false,
	Usable:      false,
	Unbreakable: true,
	SVG:         "",
}
