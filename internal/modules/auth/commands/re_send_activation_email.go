package commands

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/eskrenkovic/mediator-go"
	"net/http"
	"time"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/domain"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"
	"github.com/jmoiron/sqlx"

	"github.com/google/uuid"
)

type ReSendActivationEmailCommand struct {
	UserID uuid.UUID `json:"user_id"`
}

func (c ReSendActivationEmailCommand) Validate() error {
	if c.UserID == uuid.Nil {
		return fmt.Errorf("invalid UserID: %s", c.UserID)
	}

	return nil
}

func HandleReSendConfirmationEmail(m *mediator.Mediator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		command, err := core.RequestBody[ReSendActivationEmailCommand](r)
		if err != nil {
			core.WriteBadRequest(w, r, err)
			return
		}

		if _, err := mediator.Send[ReSendActivationEmailCommand, core.Unit](m, r.Context(), command); err != nil {
			core.WriteCommandError(w, r, err)
			return
		}

		core.WriteOK(w, r, nil)
	}
}

type ReSendActivationEmailCommandHandler struct {
	db          *sqlx.DB
	emailClient *core.EmailClient
	emailSender string
}

func NewReSendActivationEmailCommandHandler(
	db *sqlx.DB,
	emailClient *core.EmailClient,
	emailSender string,
) *ReSendActivationEmailCommandHandler {
	return &ReSendActivationEmailCommandHandler{db, emailClient, emailSender}
}

func (h ReSendActivationEmailCommandHandler) Handle(
	ctx context.Context,
	request ReSendActivationEmailCommand,
) (core.Unit, error) {
	const getUserQuery = "SELECT * FROM auth.user WHERE id = $1;"

	var user domain.User
	if err := h.db.GetContext(ctx, &user, getUserQuery, request.UserID); err != nil {
		return core.Unit{}, core.NewCommandError(400, err)
	}

	activationCode, err := domain.CreateRegistrationActivationCode(user, 7*24*time.Hour, sha256.New())
	if err != nil {
		return core.Unit{}, core.NewCommandError(400, err)
	}

	// TODO: should this be moved into the domain.CreateRegistrationActivationCode func?
	user.SecurityStamp = uuid.New()

	nowUTC := time.Now().UTC()
	activationCode.SentAt = &nowUTC

	err = core.Tx(ctx, h.db, func(ctx context.Context, tx *sqlx.Tx) error {
		const updateUserStmt = `
			UPDATE
				auth.user
			SET
				security_stamp = :security_stamp
			WHERE
				id = :id;`

		if _, err := tx.NamedExecContext(ctx, updateUserStmt, user); err != nil {
			return err
		}

		const activationCodeStmt = `
			INSERT INTO
				auth.activation_code (user_id, security_stamp, expires_at, sent_at, token, used)
			VALUES
				(:user_id, :security_stamp, :expires_at, :sent_at, :token, :used);`

		_, err = h.db.NamedExecContext(ctx, activationCodeStmt, activationCode)
		return err
	})
	if err != nil {
		return core.Unit{}, core.NewCommandError(500, err)
	}

	email := domain.RegistrationActivationEmail(user, h.emailSender)
	if err := h.emailClient.Send(email); err != nil {
		return core.Unit{}, core.NewCommandError(500, err)
	}

	return core.Unit{}, nil
}
