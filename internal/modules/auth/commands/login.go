package commands

import (
	"context"
	"fmt"
	"github.com/eskrenkovic/mediator-go"
	"net/http"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/domain"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
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

		sessionID, err := mediator.Send[LoginCommand, uuid.UUID](m, r.Context(), command)
		if err != nil {
			core.WriteCommandError(w, r, err)
			return
		}

		sessionCookie := http.Cookie{
			Name:  "chess-session",
			Value: sessionID.String(),
		}

		http.SetCookie(w, &sessionCookie)
		core.WriteOK(w, r, nil)
	}
}

type LoginCommandHandler struct {
	db             *sqlx.DB
	passwordHasher domain.PasswordHasher
}

func NewLoginCommandHandler(db *sqlx.DB, passwordHasher domain.PasswordHasher) *LoginCommandHandler {
	return &LoginCommandHandler{db, passwordHasher}
}

func (h *LoginCommandHandler) Handle(ctx context.Context, request LoginCommand) (uuid.UUID, error) {
	const stmt = `
		SELECT
			*
		FROM
			auth.user
		WHERE
			email = $1;`

	var user domain.User
	if err := h.db.GetContext(ctx, &user, stmt, request.Email); err != nil {
		// TODO: handle not found
		return uuid.Nil, core.NewCommandError(500, err)
	}

	session, authErr := user.Authenticate(request.Password, h.passwordHasher)
	err := core.Tx(ctx, h.db, func(ctx context.Context, tx *sqlx.Tx) error {
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

		if _, err := tx.NamedExecContext(ctx, updateStmt, user); err != nil {
			return core.NewCommandError(500, err, core.WithReason("failed to authenticate user"))
		}

		if authErr != nil {
			const sessionStmt = `
			INSERT INTO 
			    auth.session 
			VALUES 
			    (id, user_id, expires_at);`

			if _, err := tx.NamedExecContext(ctx, sessionStmt, session); err != nil {
				return core.NewCommandError(500, err, core.WithReason("failed to create session"))
			}
		}

		return nil
	})

	if authErr != nil {
		err = core.NewCommandError(400, authErr, core.WithReason("failed to authenticate user"))
	}

	return uuid.Nil, err
}
