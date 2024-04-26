package snailrace

import "errors"

var (
	ErrInvalidAction     = errors.New("invalid action")
	ErrInvalidParameters = errors.New("invalid parameters")
	ErrRaceNotFound      = errors.New("race not found")
	ErrInvalidRaceState  = errors.New("invalid race state")
	ErrSnailNotFound     = errors.New("snail not found")
	ErrNotSnailOwner     = errors.New("not the owner of the snail")
	ErrAlreadyJoined     = errors.New("already joined the race")
)
