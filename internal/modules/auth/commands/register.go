package commands

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/domain"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"

	"github.com/jmoiron/sqlx"
)

type RegisterCommand struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (c RegisterCommand) Validate() error {
	if c.Username == "" {
		return fmt.Errorf("invalid Username: '%s'", c.Username)
	}

	if c.Password == "" {
		return fmt.Errorf("invalid Password: '%s'", c.Password)
	}

	if c.Email == "" {
		return fmt.Errorf("invalid Email: '%s'", c.Email)
	}

	return nil
}

type RegisterCommandHandler struct {
	db             *sqlx.DB
	passwordHasher domain.PasswordHasher
}

func NewRegisterCommandHandler(db *sqlx.DB, passwordHasher domain.PasswordHasher) *RegisterCommandHandler {
	return &RegisterCommandHandler{db, passwordHasher}
}

func (h *RegisterCommandHandler) Handle(ctx context.Context, request RegisterCommand) (core.Unit, error) {
	var count int
	const existingUserQuery = `
		SELECT
			count(id)
		FROM
			auth.user
		WHERE
			username = $1 OR email = $2;`

	if err := h.db.GetContext(ctx, &count, existingUserQuery, request.Username, request.Email); err != nil {
		return core.Unit{}, core.NewCommandError(500, err, "failed to reach database")
	}

	if count > 0 {
		// Just return ok if the user already exists. If it's a valid request,
		// the user will check their email.
		return core.Unit{}, nil
	}

	user, err := domain.RegisterUser(request.Username, request.Email, request.Password, h.passwordHasher)
	if err != nil {
		return core.Unit{}, core.NewCommandError(400, err, "user registration failed")
	}

	activationCode, err := domain.CreateRegistrationActivationCode(user, 7*24*time.Hour, sha256.New())
	if err != nil {
		return core.Unit{}, core.NewCommandError(500, err, "failed to create new user entry")
	}

	err = core.Tx(ctx, h.db, func(ctx context.Context, tx *sqlx.Tx) error {
		const stmt = `
			INSERT INTO
				auth.user (id, security_stamp, username, email, password_hash)
			VALUES
				(:id, :security_stamp, :username, :email, :password_hash);`

		if _, err := h.db.NamedExecContext(ctx, stmt, user); err != nil {
			return err
		}

		const activationCodeStmt = `
			INSERT INTO
				auth.activation_code (user_id, security_stamp, expires_at, sent_at, token, used)
			VALUES
				(:user_id, :security_stamp, :expires_at, :sent_at, :token, :used);`

		_, err := h.db.NamedExecContext(ctx, activationCodeStmt, activationCode)
		return err
	})

	if err != nil {
		return core.Unit{}, core.NewCommandError(500, err, "failed to create new user entry")
	}

	return core.Unit{}, nil
}
