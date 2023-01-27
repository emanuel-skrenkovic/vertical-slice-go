package gamesession

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type CreateSessionCommand struct {
	OwnerID uuid.UUID
	Name    string
}

func (c CreateSessionCommand) Validate() error {
	if c.OwnerID == uuid.Nil {
		return fmt.Errorf("invalid OwnerID - '%s'", c.OwnerID.String())
	}

	if c.Name == "" {
		return fmt.Errorf("invalid Name - '%s'", c.Name)
	}

	return nil
}

type CreateSessionResponse struct {
	SessionID string
}

type CreateSessionCommandHandler struct {
	db *sqlx.DB
}

func NewCreateSessionCommandHandler(db *sqlx.DB) *CreateSessionCommandHandler {
	return &CreateSessionCommandHandler{db}
}

func (h *CreateSessionCommandHandler) Handle(
	ctx context.Context,
	request CreateSessionCommand,
) (CreateSessionResponse, error) {
	session := Session{
		ID:      uuid.New().String(),
		OwnerID: request.OwnerID,
	}

	const stmt = `
		INSERT INTO
			game_session(id, owner_id, name)
		VALUES
			(:id, :owner_id, :name);`

	if _, err := h.db.NamedExecContext(ctx, stmt, session); err != nil {
		return CreateSessionResponse{}, err
	}

	return CreateSessionResponse{SessionID: session.ID}, nil
}
