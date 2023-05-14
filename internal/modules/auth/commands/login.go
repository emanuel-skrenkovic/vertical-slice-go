package commands

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/domain"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"

	"github.com/eskrenkovic/mediator-go"
	tql "github.com/eskrenkovic/typeql"
)

type LoginCommand struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c LoginCommand) Validate() error {
	if c.Email == "" {
		return fmt.Errorf("invalid email: '%s'", c.Email)
	}

	if c.Password == "" {
		return fmt.Errorf("invalid password")
	}

	return nil
}

func HandleLogin(m *mediator.Mediator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: session cookie
		command, err := core.RequestBody[LoginCommand](r)
		if err != nil {
			core.WriteBadRequest(w, r, err)
			return
		}

		session, err := mediator.Send[LoginCommand, domain.Session](m, r.Context(), command)
		if err != nil {
			core.WriteCommandError(w, r, err)
			return
		}

		sessionCookie := http.Cookie{
			Name:    "chess-session",
			Value:   session.ID.String(),
			Path:    "/",
			Expires: session.ExpiresAtUTC, // TODO: does this need to be local time?
		}

		http.SetCookie(w, &sessionCookie)
		core.WriteOK(w, r, nil)
	}
}

type LoginCommandHandler struct {
	db             *sql.DB
	passwordHasher domain.PasswordHasher
}

func NewLoginCommandHandler(db *sql.DB, passwordHasher domain.PasswordHasher) *LoginCommandHandler {
	return &LoginCommandHandler{db, passwordHasher}
}

func (h *LoginCommandHandler) Handle(ctx context.Context, request LoginCommand) (domain.Session, error) {
	const stmt = `
		SELECT
			*
		FROM
			auth.user
		WHERE
			email = $1;`
	user, err := tql.QueryFirst[domain.User](ctx, h.db, stmt, request.Email)
	if err != nil {
		return domain.Session{}, core.NewCommandError(500, err)
	}

	session, authErr := user.Authenticate(request.Password, h.passwordHasher)

	txFn := func(ctx context.Context, tx *sql.Tx) error {
		// Regardless of the auth result, save the user.
		// In case it logged in successfully, the unsuccessful attempts count
		// needs to be reset to 0.
		const updateStmt = `
			UPDATE
				auth.user
			SET
				locked                      = :locked,
				unsuccessful_login_attempts = :unsuccessful_login_attempts,
				security_stamp              = :security_stamp
			WHERE
				email = :email;` // TODO: old security stamp
		if _, err := tql.Exec(ctx, tx, updateStmt, user); err != nil {
			return core.NewCommandError(500, err, core.WithReason("failed to authenticate user"))
		}

		// Only create a session if there is no auth error.
		if authErr != nil {
			return nil
		}

		const sessionStmt = `
			INSERT INTO auth.session 
				(id, user_id, expires_at, created_at, updated_at)
			VALUES 
				(:id, :user_id, :expires_at, :created_at, :updated_at);`
		if _, err := tql.Exec(ctx, tx, sessionStmt, session); err != nil {
			return core.NewCommandError(500, err, core.WithReason("failed to create session"))
		}

		return nil
	}

	if err := core.Tx(ctx, h.db, txFn); err != nil {
		return domain.Session{}, core.NewCommandError(400, authErr, core.WithReason("failed to authenticate user"))
	}

	if authErr != nil {
		return domain.Session{}, core.NewCommandError(400, authErr, core.WithReason("failed to authenticate user"))
	}

	return session, nil
}
