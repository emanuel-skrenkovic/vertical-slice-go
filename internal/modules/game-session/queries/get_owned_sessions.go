package queries

import (
	"context"
	"database/sql"
	"fmt"

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

	sessions, err := tql.Query[domain.Session](ctx, h.db, query, request.OwnerID)
	if err != nil {
		return []domain.Session{}, err
	}

	return sessions, nil
}
