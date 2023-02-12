package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"
	"github.com/google/uuid"

	"github.com/jmoiron/sqlx"
)

type CloseSessionCommand struct {
	UserID    uuid.UUID
	SessionID string
}

func (c CloseSessionCommand) Validate() error {
	if c.SessionID == "" {
		return fmt.Errorf("invalid SessionID - '%s'", c.SessionID)
	}

	return nil
}

type CloseSessionCommandHandler struct {
	db *sqlx.DB
}

func NewCloseSessionCommandHandler(db *sqlx.DB) *CloseSessionCommandHandler {
	return &CloseSessionCommandHandler{db}
}

func (h *CloseSessionCommandHandler) Handle(ctx context.Context, request CloseSessionCommand) (core.Unit, error) {
	const stmt = `
		UPDATE
			game_session
		SET
			active = false
		WHERE
			id = $1 AND owner_id == $2;`

	_, err := h.db.ExecContext(ctx, stmt, request.SessionID, request.UserID)
	switch {
	case err != nil && errors.Is(err, sql.ErrNoRows):
		return core.Unit{}, core.NewCommandError(404, err)
	case err != nil:
		return core.Unit{}, core.NewCommandError(500, err)
	}

	return core.Unit{}, nil
}
