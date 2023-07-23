package commands

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/eskrenkovic/mediator-go"
	"github.com/eskrenkovic/tql"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/game-session/domain"
	"github.com/go-chi/chi"

	"github.com/google/uuid"
)

type CreateSessionInvitationCommand struct {
	SessionID string
	InviterID uuid.UUID
	InviteeID uuid.UUID
}

func (c CreateSessionInvitationCommand) Validate() error {
	if c.SessionID == "" {
		return fmt.Errorf("invalid SessionID - '%s'", c.SessionID)
	}

	if c.InviterID == uuid.Nil {
		return fmt.Errorf("invalid InviterID - '%s'", c.InviterID)
	}

	if c.InviteeID == uuid.Nil {
		return fmt.Errorf("invalid InviteeID - '%s'", c.InviteeID)
	}

	return nil
}

func HandleCreateSessionInvitation(w http.ResponseWriter, r *http.Request) {
	command, err := core.RequestBody[CreateSessionInvitationCommand](r)
	if err != nil {
		core.WriteBadRequest(w, r, err)
	}
	command.SessionID = chi.URLParam(r, "id")

	_, err = mediator.Send[CreateSessionInvitationCommand, core.Unit](r.Context(), command)
	if err != nil {
		core.WriteCommandError(w, r, err)
		return
	}

	core.WriteOK(w, r, nil)
}

type CreateSessionInvitationCommandHandler struct {
	db *sql.DB
}

func NewCreateSessionInvitationCommandHandler(db *sql.DB) *CreateSessionInvitationCommandHandler {
	return &CreateSessionInvitationCommandHandler{db}
}

func (h *CreateSessionInvitationCommandHandler) Handle(
	ctx context.Context,
	request CreateSessionInvitationCommand,
) (core.Unit, error) {
	invitation := domain.SessionInvitation{
		ID:        uuid.New(),
		SessionID: request.SessionID,
		InviterID: request.InviterID,
		InviteeID: request.InviteeID,
	}

	const stmt = `
		INSERT INTO
			session_invitations (id, session_id, inviter_id, invitee_id, created_at)
		VALUES
			(:id, :session_id, :inviter_id, :invitee_id);`
	_, err := tql.Exec(ctx, h.db, stmt, invitation)
	return core.Unit{}, err
}
