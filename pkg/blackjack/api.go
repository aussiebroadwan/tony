package blackjack

func Running() bool {
	dealer.mu.Lock()
	defer dealer.mu.Unlock()

	return dealer.Stage != IdleStage
}

// Host initialises and starts a new game of Blackjack. It requires a  callback
// function that is invoked on game state changes, which can be used to update
// clients. It returns an error if a game is already in progress.
func Host(callback func(stage GameStage, state GameState, messageId, channelId string), messageId, channelId string) error {
	dealer.mu.Lock()
	defer dealer.mu.Unlock()

	if dealer.Stage != IdleStage && dealer.Stage != FinishedStage {
		return ErrDealerBusy
	}

	if callback == nil || messageId == "" || channelId == "" {
		return ErrInvalidAction
	}

	// Initialise a new game state
	dealer.State = newState()
	dealer.action = make(chan int)
	dealer.messageId = messageId
	dealer.channelId = channelId
	dealer.Stage = JoinStage
	dealer.onStateChange = callback

	go executeGameLoop() // Start the game loop in a new goroutine

	return nil
}

// Join attempts to add a player to the game during the joining phase. It takes
// a user Discord ID and the bet amount as parameters and returns an error if
// the player cannot join.
func Join(userId string, bet int64) error {
	dealer.mu.Lock()
	defer dealer.mu.Unlock()

	// Check if the game is in the joining phase
	if dealer.Stage != JoinStage {
		return ErrInvalidAction
	}

	// Check if the player has already joined
	for _, u := range dealer.State.Users {
		if u.Id == userId {
			return ErrAlreadyJoined
		}
	}

	// Check if the maximum number of players has been reached
	if len(dealer.State.Users) >= MaxPlayers {
		return ErrMaxPlayers
	}

	// Add the new player to the game
	dealer.State.Users = append(dealer.State.Users, User{
		Id:        userId,
		Hand:      make([]Card, 0),
		Bet:       bet,
		Blackjack: false,
	})

	dealer.commitState()
	return nil
}

// Hit deals another card to the player requesting it and checks if they bust.
// It returns an error if it's not the player's turn or the game stage is
// incorrect.
func Hit(userId string) error {
	dealer.mu.Lock()
	defer dealer.mu.Unlock()

	if dealer.Stage != RoundStage {
		return ErrInvalidAction
	}

	for i, user := range dealer.State.Users {
		if user.Id == userId {
			if i != dealer.State.PlayerTurn {
				return ErrPlayerTurn
			}

			// Deal a card to the player and update the game state
			user.Hand = append(user.Hand, dealer.State.Shoe.Draw())
			dealer.State.Users[i] = user

			// Check if the player busts or reaches 21
			if user.Hand.Score() >= MaximumHandScore {
				dealer.State.PlayerTurn++
			}

			dealer.commitState()
			dealer.action <- 1 // Notify the game loop that an action has been taken
			return nil
		}
	}

	return ErrPlayerNotFound
}

// Stand marks the player's turn as complete and advances the game to the next
// player. It returns an error if it's not the player's turn or if the stage is
// not correct for standing.
func Stand(userId string) error {
	dealer.mu.Lock()
	defer dealer.mu.Unlock()

	if dealer.Stage != RoundStage {
		return ErrInvalidAction
	}

	for i, u := range dealer.State.Users {
		if u.Id == userId {
			if i != dealer.State.PlayerTurn {
				return ErrPlayerTurn
			}

			// Advance to the next player
			dealer.State.PlayerTurn++
			dealer.commitState()
			dealer.action <- 1 // Notify the game loop that an action has been taken
			return nil
		}
	}

	return ErrPlayerNotFound
}
