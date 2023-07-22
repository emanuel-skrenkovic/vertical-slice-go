package core

import (
	"context"

	"github.com/google/uuid"
)

type ContextKey string

const SessionContextKey ContextKey = "session"

type ContextSession struct {
	UserID uuid.UUID
}

func Session(ctx context.Context) ContextSession {
	rawVal := ctx.Value(SessionContextKey)

	if rawVal == nil {
		// TODO: this is sucks
		return ContextSession{}
	}

	session, ok := rawVal.(ContextSession)
	if !ok {
		return ContextSession{}
	}

	return session
}
