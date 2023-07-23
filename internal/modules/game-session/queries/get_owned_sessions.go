package queries

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/eskrenkovic/mediator-go"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/game-session/domain"

	"github.com/eskrenkovic/tql"
	"github.com/google/uuid"
)

type GetOwnedSessionsQuery struct {
	OwnerID uuid.UUID
}

func (q GetOwnedSessionsQuery) Validate() error {
	if q.OwnerID == uuid.Nil {
		return fmt.Errorf("invalid OwnerID - %s", q.OwnerID.String())
	}

	return nil
}

func HandleGetOwnedSessions(w http.ResponseWriter, r *http.Request) {
	ownerIDParam, found := r.URL.Query()["ownerId"]
	if !found {
		core.WriteBadRequest(w, r, fmt.Errorf("missing required query param 'ownerId'"))
		return
	}

	ownerID, err := uuid.Parse(ownerIDParam[0])
	if err != nil {
		core.WriteBadRequest(w, r, fmt.Errorf("invalid format for query param 'ownerId'"))
		return
	}

	response, err := mediator.Send[GetOwnedSessionsQuery, []domain.Session](
		r.Context(),
		GetOwnedSessionsQuery{OwnerID: ownerID},
	)
	if err != nil {
		core.WriteCommandError(w, r, err)
		return
	}

	core.WriteOK(w, r, response)
}

type GetOwnedSessionsQueryHandler struct {
	db *sql.DB
}

func NewGetOwnedSessionsQueryHandler(db *sql.DB) *GetOwnedSessionsQueryHandler {
	return &GetOwnedSessionsQueryHandler{db}
}

func (h *GetOwnedSessionsQueryHandler) Handle(
	ctx context.Context,
	request GetOwnedSessionsQuery,
) ([]domain.Session, error) {
	const query = `
		SELECT
			*
		FROM
			game_session
		WHERE
			owner_id = $1;`
	return tql.Query[domain.Session](ctx, h.db, query, request.OwnerID)
}
