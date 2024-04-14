package blackjack

import "errors"

var (
	ErrDealerBusy     = errors.New("dealer is busy")
	ErrAlreadyJoined  = errors.New("user already joined")
	ErrMaxPlayers     = errors.New("max players reached")
	ErrInvalidAction  = errors.New("invalid action")
	ErrPlayerTurn     = errors.New("not player's turn")
	ErrPlayerNotFound = errors.New("player not found")
)
