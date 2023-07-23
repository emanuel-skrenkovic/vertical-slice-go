package commands

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"github.com/eskrenkovic/mediator-go"
	"net/http"
	"time"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/domain"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"

	"github.com/eskrenkovic/tql"
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

func HandleRegistration(w http.ResponseWriter, r *http.Request) {
	command, err := core.RequestBody[RegisterCommand](r)
	if err != nil {
		core.WriteBadRequest(w, r, err)
	}

	_, err = mediator.Send[RegisterCommand, core.Unit](r.Context(), command)
	if err != nil {
		core.WriteCommandError(w, r, err)
		return
	}

	core.WriteOK(w, r, nil)
}

type RegisterCommandHandler struct {
	db             *sql.DB
	passwordHasher domain.PasswordHasher
}

func NewRegisterCommandHandler(db *sql.DB, passwordHasher domain.PasswordHasher) *RegisterCommandHandler {
	return &RegisterCommandHandler{db, passwordHasher}
}

func (h *RegisterCommandHandler) Handle(ctx context.Context, request RegisterCommand) (core.Unit, error) {
	const existingUserQuery = `
		SELECT
			count(id)
		FROM
			auth.user
		WHERE
			username = $1 OR email = $2;`

	count, err := tql.QueryFirst[int](ctx, h.db, existingUserQuery, request.Username, request.Email)
	if err != nil {
		return core.Unit{}, core.NewCommandError(500, err)
	}

	// Just return ok if the user already exists. If it's a valid request,
	// the user will check their email.
	if count > 0 {
		return core.Unit{}, nil
	}

	user, err := domain.RegisterUser(request.Username, request.Email, request.Password, h.passwordHasher)
	if err != nil {
		return core.Unit{}, core.NewCommandError(400, err)
	}

	// TODO: pull duration from configuration
	activationCode, err := domain.CreateRegistrationActivationCode(user, 7*24*time.Hour, sha256.New())
	if err != nil {
		return core.Unit{}, core.NewCommandError(500, err)
	}

	err = core.Tx(ctx, h.db, func(ctx context.Context, tx *sql.Tx) error {
		const stmt = `
			INSERT INTO
				auth.user (id, security_stamp, username, email, password_hash)
			VALUES
				(:id, :security_stamp, :username, :email, :password_hash);`

		if _, err := tql.Exec(ctx, tx, stmt, user); err != nil {
			return err
		}

		const activationCodeStmt = `
			INSERT INTO
				auth.activation_code (user_id, security_stamp, expires_at, sent_at, token, used)
			VALUES
				(:user_id, :security_stamp, :expires_at, :sent_at, :token, :used);`

		_, err := tql.Exec(ctx, tx, activationCodeStmt, activationCode)
		return err
	})

	// Since changes need to be stored regardless of the auth result,
	// set the error to be the auth result even if the save fails
	// so the user sees that instead.
	if err != nil {
		return core.Unit{}, core.NewCommandError(500, err)
	}

	return core.Unit{}, nil
}
