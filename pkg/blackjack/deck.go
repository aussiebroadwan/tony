package blackjack

import "math/rand"

var Deck = []Card{
	CardAceSpades, CardTwoSpades, CardThreeSpades, CardFourSpades, CardFiveSpades, CardSixSpades, CardSevenSpades, CardEightSpades, CardNineSpades, CardTenSpades, CardJackSpades, CardQueenSpades, CardKingSpades,
	CardAceHearts, CardTwoHearts, CardThreeHearts, CardFourHearts, CardFiveHearts, CardSixHearts, CardSevenHearts, CardEightHearts, CardNineHearts, CardTenHearts, CardJackHearts, CardQueenHearts, CardKingHearts,
	CardAceDiamonds, CardTwoDiamonds, CardThreeDiamonds, CardFourDiamonds, CardFiveDiamonds, CardSixDiamonds, CardSevenDiamonds, CardEightDiamonds, CardNineDiamonds, CardTenDiamonds, CardJackDiamonds, CardQueenDiamonds, CardKingDiamonds,
	CardAceClubs, CardTwoClubs, CardThreeClubs, CardFourClubs, CardFiveClubs, CardSixClubs, CardSevenClubs, CardEightClubs, CardNineClubs, CardTenClubs, CardJackClubs, CardQueenClubs, CardKingClubs,
}

type Shoe []Card

func NewShoe(decks int) Shoe {
	var shoe Shoe
	for i := 0; i < decks; i++ {
		shoe = append(shoe, Deck...)
	}
	return shoe
}

func (s *Shoe) Shuffle() {
	shoe := *s
	for i := range shoe {
		j := i + rand.Intn(len(shoe)-i)
		shoe[i], shoe[j] = shoe[j], shoe[i]
	}
}

func (s *Shoe) Draw() Card {
	shoe := *s
	card := shoe[0]
	*s = shoe[1:]
	return card
}

type Hand []Card

func (h Hand) Score() int {
	score := 0
	aces := 0
	for _, card := range h {
		score += card.Score
		if card.Rank == RankAce {
			aces++
		}
	}
	for aces > 0 && score > 21 {
		score -= 10
		aces--
	}
	return score
}
