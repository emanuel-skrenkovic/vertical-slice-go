package commands

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"path"

	"github.com/eskrenkovic/mediator-go"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/game-session/domain"

	"github.com/eskrenkovic/tql"
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

func HandleCreateGameSession(w http.ResponseWriter, r *http.Request) {
	command, err := core.RequestBody[CreateSessionCommand](r)
	if err != nil {
		core.WriteBadRequest(w, r, err)
		return
	}

	response, err := mediator.Send[CreateSessionCommand, CreateSessionResponse](
		r.Context(),
		command,
	)
	if err != nil {
		core.WriteCommandError(w, r, err)
		return
	}

	location := path.Join(r.Host, "game-sessions", response.SessionID)
	core.WriteCreated(w, r, location)
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
		ID:      uuid.NewString(),
		OwnerID: request.OwnerID,
	}

	const stmt = `
		INSERT INTO
			game_session (id, owner_id, name)
		VALUES
			(:id, :owner_id, :name);`
	if _, err := tql.Exec(ctx, h.db, stmt, session); err != nil {
		return CreateSessionResponse{}, err
	}

	return CreateSessionResponse{SessionID: session.ID}, nil
}
