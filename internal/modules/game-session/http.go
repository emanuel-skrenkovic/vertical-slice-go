package gamesession

import (
	"fmt"
	"net/http"
	"path"

	"github.com/eskrenkovic/mediator-go"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"

	"github.com/google/uuid"
)

type GameSessionHTTPHandler struct {
	m *mediator.Mediator
}

func NewGameSessionHTTPHandler(m *mediator.Mediator) *GameSessionHTTPHandler {
	return &GameSessionHTTPHandler{m}
}

func (h *GameSessionHTTPHandler) HandleCreateGameSession(w http.ResponseWriter, r *http.Request) {
	command, err := core.RequestBody[CreateSessionCommand](r)
	if err != nil {
		core.WriteBadRequest(w, r, err)
		return
	}

	response, err := mediator.Send[CreateSessionCommand, CreateSessionResponse](h.m, r.Context(), command)
	if err != nil {
		// TODO: don't like this at all. Needs to be a simple function call or a decorator solution.
		statusCode := 500
		if commandErr, ok := err.(core.CommandError); ok {
			statusCode = commandErr.StatusCode
		}
		core.WriteResponse(w, r, statusCode, err)
		return
	}

	location := path.Join(r.Host, "game-sessions", response.SessionID)
	core.WriteCreated(w, r, location)
}

func (h *GameSessionHTTPHandler) HandleGetOwnedSessions(w http.ResponseWriter, r *http.Request) {
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

	response, err := mediator.Send[GetOwnedSessionsQuery, []Session](
		h.m,
		r.Context(),
		GetOwnedSessionsQuery{OwnerID: ownerID},
	)
	if err != nil {
		// TODO: don't like this at all. Needs to be a simple function call or a decorator solution.
		statusCode := 500
		if commandErr, ok := err.(core.CommandError); ok {
			statusCode = commandErr.StatusCode
		}
		core.WriteResponse(w, r, statusCode, err)
		return
	}

	core.WriteOK(w, r, response)
}
