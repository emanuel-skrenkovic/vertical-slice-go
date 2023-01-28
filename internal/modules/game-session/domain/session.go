package domain

import "github.com/google/uuid"

type Session struct {
	ID        string    `db:"id"`
	OwnerID   uuid.UUID `db:"owner_id"`
	Player1ID uuid.UUID `db:"player_1_id"`
	Player2ID uuid.UUID `db:"player_2_id"`
	GameID    uuid.UUID `db:"game_id"`
	Active    bool      `db:"active"`
	Name      string    `db:"name"`
}

// Create session +
// See your open sessions +
// See sessions you are invited to
// Join session
// Close session (if you created it)
