package blackjack

import "fmt"

const (
	// Suits for the cards
	SuitSpades   = "♠"
	SuitHearts   = "♥"
	SuitDiamonds = "♦"
	SuitClubs    = "♣"

	// Ranks
	RankAce   = "A"
	RankTwo   = "2"
	RankThree = "3"
	RankFour  = "4"
	RankFive  = "5"
	RankSix   = "6"
	RankSeven = "7"
	RankEight = "8"
	RankNine  = "9"
	RankTen   = "10"
	RankJack  = "J"
	RankQueen = "Q"
	RankKing  = "K"
)

type Card struct {
	Score int
	Rank  string
	Suit  string
}

func (c Card) String() string {
	return fmt.Sprintf("%s %s", c.Rank, c.Suit)
}

var (
	// Spades
	CardAceSpades   = Card{Score: 11, Rank: RankAce, Suit: SuitSpades}
	CardTwoSpades   = Card{Score: 2, Rank: RankTwo, Suit: SuitSpades}
	CardThreeSpades = Card{Score: 3, Rank: RankThree, Suit: SuitSpades}
	CardFourSpades  = Card{Score: 4, Rank: RankFour, Suit: SuitSpades}
	CardFiveSpades  = Card{Score: 5, Rank: RankFive, Suit: SuitSpades}
	CardSixSpades   = Card{Score: 6, Rank: RankSix, Suit: SuitSpades}
	CardSevenSpades = Card{Score: 7, Rank: RankSeven, Suit: SuitSpades}
	CardEightSpades = Card{Score: 8, Rank: RankEight, Suit: SuitSpades}
	CardNineSpades  = Card{Score: 9, Rank: RankNine, Suit: SuitSpades}
	CardTenSpades   = Card{Score: 10, Rank: RankTen, Suit: SuitSpades}
	CardJackSpades  = Card{Score: 10, Rank: RankJack, Suit: SuitSpades}
	CardQueenSpades = Card{Score: 10, Rank: RankQueen, Suit: SuitSpades}
	CardKingSpades  = Card{Score: 10, Rank: RankKing, Suit: SuitSpades}

	// Hearts
	CardAceHearts   = Card{Score: 11, Rank: RankAce, Suit: SuitHearts}
	CardTwoHearts   = Card{Score: 2, Rank: RankTwo, Suit: SuitHearts}
	CardThreeHearts = Card{Score: 3, Rank: RankThree, Suit: SuitHearts}
	CardFourHearts  = Card{Score: 4, Rank: RankFour, Suit: SuitHearts}
	CardFiveHearts  = Card{Score: 5, Rank: RankFive, Suit: SuitHearts}
	CardSixHearts   = Card{Score: 6, Rank: RankSix, Suit: SuitHearts}
	CardSevenHearts = Card{Score: 7, Rank: RankSeven, Suit: SuitHearts}
	CardEightHearts = Card{Score: 8, Rank: RankEight, Suit: SuitHearts}
	CardNineHearts  = Card{Score: 9, Rank: RankNine, Suit: SuitHearts}
	CardTenHearts   = Card{Score: 10, Rank: RankTen, Suit: SuitHearts}
	CardJackHearts  = Card{Score: 10, Rank: RankJack, Suit: SuitHearts}
	CardQueenHearts = Card{Score: 10, Rank: RankQueen, Suit: SuitHearts}
	CardKingHearts  = Card{Score: 10, Rank: RankKing, Suit: SuitHearts}

	// Diamonds
	CardAceDiamonds   = Card{Score: 11, Rank: RankAce, Suit: SuitDiamonds}
	CardTwoDiamonds   = Card{Score: 2, Rank: RankTwo, Suit: SuitDiamonds}
	CardThreeDiamonds = Card{Score: 3, Rank: RankThree, Suit: SuitDiamonds}
	CardFourDiamonds  = Card{Score: 4, Rank: RankFour, Suit: SuitDiamonds}
	CardFiveDiamonds  = Card{Score: 5, Rank: RankFive, Suit: SuitDiamonds}
	CardSixDiamonds   = Card{Score: 6, Rank: RankSix, Suit: SuitDiamonds}
	CardSevenDiamonds = Card{Score: 7, Rank: RankSeven, Suit: SuitDiamonds}
	CardEightDiamonds = Card{Score: 8, Rank: RankEight, Suit: SuitDiamonds}
	CardNineDiamonds  = Card{Score: 9, Rank: RankNine, Suit: SuitDiamonds}
	CardTenDiamonds   = Card{Score: 10, Rank: RankTen, Suit: SuitDiamonds}
	CardJackDiamonds  = Card{Score: 10, Rank: RankJack, Suit: SuitDiamonds}
	CardQueenDiamonds = Card{Score: 10, Rank: RankQueen, Suit: SuitDiamonds}
	CardKingDiamonds  = Card{Score: 10, Rank: RankKing, Suit: SuitDiamonds}

	// Clubs
	CardAceClubs   = Card{Score: 11, Rank: RankAce, Suit: SuitClubs}
	CardTwoClubs   = Card{Score: 2, Rank: RankTwo, Suit: SuitClubs}
	CardThreeClubs = Card{Score: 3, Rank: RankThree, Suit: SuitClubs}
	CardFourClubs  = Card{Score: 4, Rank: RankFour, Suit: SuitClubs}
	CardFiveClubs  = Card{Score: 5, Rank: RankFive, Suit: SuitClubs}
	CardSixClubs   = Card{Score: 6, Rank: RankSix, Suit: SuitClubs}
	CardSevenClubs = Card{Score: 7, Rank: RankSeven, Suit: SuitClubs}
	CardEightClubs = Card{Score: 8, Rank: RankEight, Suit: SuitClubs}
	CardNineClubs  = Card{Score: 9, Rank: RankNine, Suit: SuitClubs}
	CardTenClubs   = Card{Score: 10, Rank: RankTen, Suit: SuitClubs}
	CardJackClubs  = Card{Score: 10, Rank: RankJack, Suit: SuitClubs}
	CardQueenClubs = Card{Score: 10, Rank: RankQueen, Suit: SuitClubs}
	CardKingClubs  = Card{Score: 10, Rank: RankKing, Suit: SuitClubs}
)
