package gamesession

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
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
	db *sqlx.DB
}

func NewGetOwnedSessionsQueryHandler(db *sqlx.DB) *GetOwnedSessionsQueryHandler {
	return &GetOwnedSessionsQueryHandler{db}
}

func (h *GetOwnedSessionsQueryHandler) Handle(
	ctx context.Context,
	request GetOwnedSessionsQuery,
) ([]Session, error) {
	const query = `
		SELECT
			*
		FROM
			game_session
		WHERE
			owner_id = $1;`

	sessions := make([]Session, 0)
	if err := h.db.SelectContext(ctx, &sessions, query, request.OwnerID); err != nil {
		return nil, err
	}

	return sessions, nil
}
