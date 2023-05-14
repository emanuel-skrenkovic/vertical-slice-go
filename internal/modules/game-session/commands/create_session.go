package commands

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/eskrenkovic/vertical-slice-go/internal/tql"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/game-session/domain"

	"github.com/google/uuid"
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
	db *sql.DB
}

func NewCreateSessionCommandHandler(db *sql.DB) *CreateSessionCommandHandler {
	return &CreateSessionCommandHandler{db}
}

func (h *CreateSessionCommandHandler) Handle(
	ctx context.Context,
	request CreateSessionCommand,
) (CreateSessionResponse, error) {
	session := domain.Session{
		ID:      uuid.New().String(),
		OwnerID: request.OwnerID,
	}

	const stmt = `
		INSERT INTO
			game_session(id, owner_id, name)
		VALUES
			(:id, :owner_id, :name);`

	if _, err := tql.Exec(ctx, h.db, stmt, session); err != nil {
		return CreateSessionResponse{}, err
	}

	return CreateSessionResponse{SessionID: session.ID}, nil
}
