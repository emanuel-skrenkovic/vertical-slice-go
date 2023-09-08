package commands

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"

	"github.com/eskrenkovic/mediator-go"
	"github.com/eskrenkovic/tql"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type JoinSessionCommand struct {
	PlayerID  uuid.UUID
	SessionID string
}

func (c JoinSessionCommand) Validate() error {
	if c.PlayerID == uuid.Nil {
		return fmt.Errorf("invalid PlayerID - '%s'", c.PlayerID)
	}

	if c.SessionID == "" {
		return fmt.Errorf("invalid SessionID - '%s'", c.SessionID)
	}

	return nil
}

type JoinSessionResponse struct {
	SessionURL string
}

func HandleJoinSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	command := JoinSessionCommand{
		SessionID: chi.URLParam(r, "id"),
		PlayerID:  core.Session(ctx).UserID, // you join someone else's session as the logged-in user
	}

	_, err := mediator.Send[JoinSessionCommand, core.Unit](ctx, command)
	if err != nil {
		core.WriteCommandError(w, r, err)
		return
	}

	core.WriteOK(w, r, nil)
}

type JoinSessionCommandHandler struct {
	db *sql.DB
}

func NewJoinSessionCommandHandler(db *sql.DB) *JoinSessionCommandHandler {
	return &JoinSessionCommandHandler{db}
}

func (h *JoinSessionCommandHandler) Handle(
	ctx context.Context,
	request JoinSessionCommand,
) (JoinSessionResponse, error) {
	const stmt = `
		UPDATE
			game_session
		SET
			active = true
		WHERE
			session_id = $1 AND owner_id = $2;`
	if _, err := tql.Exec(ctx, h.db, stmt, request.SessionID, request.PlayerID); err != nil {
		return JoinSessionResponse{}, core.NewCommandError(400, err)
	}

	return JoinSessionResponse{SessionURL: "TODO"}, nil
}
