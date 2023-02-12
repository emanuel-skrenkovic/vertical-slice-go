package commands

import (
	"context"
	"fmt"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/domain"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"

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

type LoginCommandHandler struct {
	db             *sqlx.DB
	passwordHasher domain.PasswordHasher
}

func NewLoginCommandHandler(db *sqlx.DB, passwordHasher domain.PasswordHasher) *LoginCommandHandler {
	return &LoginCommandHandler{db, passwordHasher}
}

func (h *LoginCommandHandler) Handle(ctx context.Context, request LoginCommand) (core.Unit, error) {
	var user domain.User

	const stmt = `
		SELECT
			*
		FROM
			auth.user
		WHERE
			email = $1;`

	if err := h.db.GetContext(ctx, &user, stmt, request.Email); err != nil {
		// TODO: handle not found
		return core.Unit{}, core.NewCommandError(500, err)
	}

	authErr := user.Authenticate(request.Password, h.passwordHasher)

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

	if _, err := h.db.NamedExecContext(ctx, updateStmt, user); err != nil {
		return core.Unit{}, core.NewCommandError(500, err, core.WithReason("failed to authenticate user"))
	}

	var errResult error
	if authErr != nil {
		errResult = core.NewCommandError(400, authErr, core.WithReason("failed to authenticate user"))
	}

	return core.Unit{}, errResult
}
