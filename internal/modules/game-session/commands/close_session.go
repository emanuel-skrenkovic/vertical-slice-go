package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"

	"github.com/eskrenkovic/mediator-go"
	"github.com/eskrenkovic/tql"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
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

func HandleCloseSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	command := CloseSessionCommand{
		SessionID: chi.URLParam(r, "id"),
		UserID:    core.Session(ctx).UserID,
	}

	_, err := mediator.Send[CloseSessionCommand, core.Unit](ctx, command)
	if err != nil {
		core.WriteCommandError(w, r, err)
		return
	}

	core.WriteOK(w, r, nil)
}

type CloseSessionCommandHandler struct {
	db *sql.DB
}

func NewCloseSessionCommandHandler(db *sql.DB) *CloseSessionCommandHandler {
	return &CloseSessionCommandHandler{db}
}

func (h *CloseSessionCommandHandler) Handle(
	ctx context.Context,
	request CloseSessionCommand,
) (core.Unit, error) {
	txFn := func(context.Context, *sql.Tx) error {
		const stmt = `
			UPDATE
				game_session
			SET
				active = false
			WHERE
				id = $1 AND owner_id == $2;`

		if _, err := tql.Exec(ctx, h.db, stmt, request.SessionID, request.UserID); err != nil {
			return err
		}

		const invitationStmt = `
			UPDATE
				session_invitations
			SET
				active = false
			WHERE
				session_id = $1;`
		_, err := tql.Exec(ctx, h.db, invitationStmt, request.SessionID)
		return err
	}

	err := core.Tx(ctx, h.db, txFn)
	switch {
	case err != nil && errors.Is(err, sql.ErrNoRows):
		return core.Unit{}, core.NewCommandError(404, err)
	case err != nil:
		return core.Unit{}, core.NewCommandError(500, err)
	}

	return core.Unit{}, nil
}
