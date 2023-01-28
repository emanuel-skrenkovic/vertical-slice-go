package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/domain"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"
	"github.com/google/uuid"

	"github.com/jmoiron/sqlx"
)

type VerifyRegistrationCommand struct {
	Token string
}

func (c VerifyRegistrationCommand) Validate() error {
	if c.Token == "" {
		return fmt.Errorf("invalid Token: '%s'", c.Token)
	}

	return nil
}

type VerifyRegistrationCommandHandler struct {
	db *sqlx.DB
}

func (h *VerifyRegistrationCommandHandler) Handle(
	ctx context.Context,
	request VerifyRegistrationCommand,
) (core.Unit, error) {
	const invalidTokenMessage = "invalid confirmation token"

	token, err := domain.ParseConfirmationToken(request.Token)
	if err != nil {
		return core.Unit{}, core.NewCommandError(400, err, invalidTokenMessage)
	}

	const stmt = `
		SELECT
			*
		FROM
			auth.user
		WHERE
			id = $1 AND security_stamp = $2;`

	var user domain.User
	if err := h.db.GetContext(ctx, &user, stmt, token.UserID, token.SecurityStamp); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return core.Unit{}, core.NewCommandError(500, err, invalidTokenMessage)
		}

		return core.Unit{}, core.NewCommandError(500, err, "failed to get user from database")
	}

	if err := domain.ValidateUserConfirmationToken(token, user); err != nil {
		return core.Unit{}, core.NewCommandError(500, err, invalidTokenMessage)
	}

	// TODO: should the security stamp be updated if the confirmation fails?

	updateParams := map[string]interface{}{
		"old_security_stamp": token.SecurityStamp,
		"new_security_stamp": uuid.New(),
	}

	const updateUserStmt = `
		UPDATE
			auth.user
		SET
			security_stamp = :new_security_stamp
			email_confirmed = true
		WHERE
			id = :user_id AND security_stamp = :old_security_stamp;`

	if _, err := h.db.ExecContext(ctx, updateUserStmt, updateParams); err != nil {
		return core.Unit{}, core.NewCommandError(500, err, "failed to store confirmed user")
	}

	return core.Unit{}, nil
}
