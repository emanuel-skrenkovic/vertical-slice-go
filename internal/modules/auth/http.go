package auth

import (
	"fmt"
	"net/http"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/commands"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"

	"github.com/eskrenkovic/mediator-go"
)

type AuthHTTPHandler struct {
	m *mediator.Mediator
}

func NewAuthHTTPHandler(m *mediator.Mediator) *AuthHTTPHandler {
	return &AuthHTTPHandler{m}
}

func (h *AuthHTTPHandler) HandleRegistration(w http.ResponseWriter, r *http.Request) {
	command, err := core.RequestBody[commands.RegisterCommand](r)
	if err != nil {
		core.WriteBadRequest(w, r, err)
	}

	if _, err = mediator.Send[commands.RegisterCommand, core.Unit](h.m, r.Context(), command); err != nil {
		core.WriteCommandError(w, r, err)
		return
	}

	core.WriteOK(w, r, nil)
}

func (h *AuthHTTPHandler) HandleVerifyRegistration(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		core.WriteBadRequest(w, r, fmt.Errorf("invalid token"))
	}

	command := commands.VerifyRegistrationCommand{Token: token}
	_, err := mediator.Send[commands.VerifyRegistrationCommand, core.Unit](h.m, r.Context(), command)
	if err != nil {
		core.WriteCommandError(w, r, err)
		return
	}

	core.WriteOK(w, r, nil)
}

func (h *AuthHTTPHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// TODO: session cookie
	command, err := core.RequestBody[commands.LoginCommand](r)
	if err != nil {
		core.WriteBadRequest(w, r, err)
		return
	}

	_, err = mediator.Send[commands.LoginCommand, core.Unit](h.m, r.Context(), command)
	if err != nil {
		core.WriteCommandError(w, r, err)
		return
	}

	sessionCookie := http.Cookie{
		Name:  "chess-session",
		Value: "HI MOM",
	}

	http.SetCookie(w, &sessionCookie)
	core.WriteOK(w, r, nil)
}

func (h *AuthHTTPHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{MaxAge: -1})
	core.WriteOK(w, r, nil)
}

func (h *AuthHTTPHandler) HandlePublishConfirmationEmails(w http.ResponseWriter, r *http.Request) {
	command := commands.ProcessActivationCodesCommand{}
	_, err := mediator.Send[commands.ProcessActivationCodesCommand, core.Unit](h.m, r.Context(), command)
	if err != nil {
		core.WriteCommandError(w, r, err)
		return
	}

	core.WriteOK(w, r, nil)
}

func (h *AuthHTTPHandler) HandleReSendConnfirmationEmail(w http.ResponseWriter, r *http.Request) {
	command, err := core.RequestBody[commands.ReSendActivationEmailCommand](r)
	if err != nil {
		core.WriteBadRequest(w, r, err)
		return
	}

	_, err = mediator.Send[commands.ReSendActivationEmailCommand, core.Unit](h.m, r.Context(), command)
	if err != nil {
		core.WriteCommandError(w, r, err)
		return
	}

	core.WriteOK(w, r, nil)
}
