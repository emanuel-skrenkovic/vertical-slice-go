package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/domain"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"

	"github.com/eskrenkovic/mediator-go"
	"github.com/eskrenkovic/tql"
	"github.com/google/uuid"
)

type VerifyRegistrationCommand struct {
	Token string `json:"token"`
}

func (c VerifyRegistrationCommand) Validate() error {
	if c.Token == "" {
		return fmt.Errorf("invalid Token: '%s'", c.Token)
	}

	return nil
}

func HandleVerifyRegistration(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		core.WriteBadRequest(w, r, fmt.Errorf("invalid token"))
	}

	command := VerifyRegistrationCommand{Token: token}
	_, err := mediator.Send[VerifyRegistrationCommand, core.Unit](r.Context(), command)
	if err != nil {
		core.WriteCommandError(w, r, err)
		return
	}

	core.WriteOK(w, r, nil)
}

type VerifyRegistrationCommandHandler struct {
	db *sql.DB
}

func NewVerifyRegistrationCommandHandler(db *sql.DB) *VerifyRegistrationCommandHandler {
	return &VerifyRegistrationCommandHandler{db}
}

func (h *VerifyRegistrationCommandHandler) Handle(
	ctx context.Context,
	request VerifyRegistrationCommand,
) (core.Unit, error) {
	const invalidTokenMessage = "invalid confirmation token"

	const getCodeQuery = `
		SELECT
			*
		FROM
			auth.activation_code
		WHERE
			token = $1;`

	activationCode, err := tql.QueryFirst[domain.ActivationCode](ctx, h.db, getCodeQuery, request.Token)
	if err != nil {
		return core.Unit{}, core.NewCommandError(400, fmt.Errorf("invalid activation code"))
	}

	const stmt = `
		SELECT
			*
		FROM
			auth.user
		WHERE
			id = $1 AND security_stamp = $2;`

	user, err := tql.QueryFirst[domain.User](ctx, h.db, stmt, activationCode.UserID, activationCode.SecurityStamp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return core.Unit{}, core.NewCommandError(500, err, core.WithReason(invalidTokenMessage))
		}
		return core.Unit{}, core.NewCommandError(500, err)
	}

	if err := domain.ValidateUserActivationCode(activationCode, user); err != nil {
		// TODO: should the security stamp be updated if the confirmation fails?
		return core.Unit{}, core.NewCommandError(400, err, core.WithReason(invalidTokenMessage))
	}

	updateParams := map[string]interface{}{
		"user_id":            user.ID,
		"old_security_stamp": activationCode.SecurityStamp,
		"new_security_stamp": uuid.New(),
	}

	err = core.Tx(ctx, h.db, func(ctx context.Context, tx *sql.Tx) error {
		const updateUserStmt = `
			UPDATE
				auth.user
			SET
				security_stamp = :new_security_stamp,
				email_confirmed = true
			WHERE
				id = :user_id AND security_stamp = :old_security_stamp;`

		if _, err := tql.Exec(ctx, tx, updateUserStmt, updateParams); err != nil {
			return err
		}

		const updateActivationCodeStmt = `
			UPDATE
				auth.activation_code
			SET
				used = true
			WHERE
				token = $1;`

		_, err := tql.Exec(ctx, tx, updateActivationCodeStmt, activationCode.Token)
		return err
	})

	if err != nil {
		return core.Unit{}, core.NewCommandError(500, err)
	}

	return core.Unit{}, nil
}
