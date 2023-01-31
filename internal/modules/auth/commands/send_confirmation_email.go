package commands

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/domain"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"

	"github.com/jmoiron/sqlx"
)

type SendConfirmationEmailCommand struct{}

type SendConfirmationEmailCommandHandler struct {
	db *sqlx.DB
}

func NewSendConfirmationEmailCommandHandler() *SendConfirmationEmailCommandHandler {
	return &SendConfirmationEmailCommandHandler{}
}

func (h *SendConfirmationEmailCommandHandler) Handle(
	ctx context.Context,
	_ SendConfirmationEmailCommand,
) (core.Unit, error) {
	const stmt = `
		SELECT
			*
		FROM
			auth.authentication_code c
		INNER JOIN
			auth.user u ON c.user_id = u.id
		WHERE
			u.email_confirmed = false AND c.expires_at > $1;`

	var codes []domain.ActivationCode
	err := h.db.SelectContext(ctx, &codes, stmt, time.Now().UTC())
	switch {
	case err != nil && errors.Is(err, sql.ErrNoRows):
		return core.Unit{}, nil
	case err != nil:
		return core.Unit{}, core.NewCommandError(500, err, "failed to get users")
	}

	// TODO: send email to each of the users.
	// Kind of want to have a 'registration' table
	// which will track the account confirmation.
	// Token would be a hash, and the confirmation mail could be re-sent.
	// https://security.stackexchange.com/questions/197004/what-should-a-verification-email-consist-of

	for _, code := range codes {
		// TODO: SEND EMAIL. Email client a prerequisite.
		_ = code
	}

	// Mark all the codes as sent. If the user did not receive the code,
	// they can click the re-send activation code button.
	// It is prefered to not send the email than to send multiple ones.

	return core.Unit{}, nil
}
