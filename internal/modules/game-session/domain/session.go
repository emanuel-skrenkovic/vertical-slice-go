package domain

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        string    `db:"id"`
	OwnerID   uuid.UUID `db:"owner_id"`
	Player1ID uuid.UUID `db:"player_1_id"`
	Player2ID uuid.UUID `db:"player_2_id"`
	GameID    uuid.UUID `db:"game_id"`
	Active    bool      `db:"active"`
	Name      string    `db:"name"`
}

type SessionInvitation struct {
	ID        uuid.UUID `db:"id"`
	SessionID string    `db:"session_id"`

	InviterID uuid.UUID `db:"inviter_id"`
	InviteeID uuid.UUID `db:"invitee_id"`

	CreatedAt time.Time `db:"created_at"`
}

// Create session +
// See your open sessions +
// See sessions you are invited to
// Join session
// Close session (if you created it)
