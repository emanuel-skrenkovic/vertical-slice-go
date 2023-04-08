package commands

import (
	"context"
	"database/sql"
	"errors"
	"github.com/eskrenkovic/mediator-go"
	"net/http"
	"time"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/domain"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/jmoiron/sqlx"
)

type EmailConfiguration struct {
	Sender string
}

type ProcessActivationCodesCommand struct{}

func HandlePublishConfirmationEmails(m *mediator.Mediator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := mediator.Send[ProcessActivationCodesCommand, core.Unit](
			m,
			r.Context(),
			ProcessActivationCodesCommand{},
		)
		if err != nil {
			core.WriteCommandError(w, r, err)
			return
		}

		core.WriteOK(w, r, nil)
	}
}

type ProcessActivationCodesCommandHandler struct {
	db          *sqlx.DB
	emailClient *core.EmailClient
	emailConfig EmailConfiguration
}

func NewProcessActivationCodesCommandHandler(
	db *sqlx.DB,
	emailClient *core.EmailClient,
	emailConfig EmailConfiguration,
) *ProcessActivationCodesCommandHandler {
	return &ProcessActivationCodesCommandHandler{db, emailClient, emailConfig}
}

func (h *ProcessActivationCodesCommandHandler) Handle(
	ctx context.Context,
	_ ProcessActivationCodesCommand,
) (core.Unit, error) {
	const stmt = `
		SELECT
			c.*
		FROM
			auth.activation_code c
		INNER JOIN
			auth.user u ON c.user_id = u.id AND u.security_stamp = c.security_stamp
		WHERE
			u.email_confirmed = false AND c.expires_at > $1;`

	var codes []domain.ActivationCode
	if err := h.db.SelectContext(ctx, &codes, stmt, time.Now().UTC()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return core.Unit{}, nil
		}
		return core.Unit{}, core.NewCommandError(500, err)
	}

	// https://security.stackexchange.com/questions/197004/what-should-a-verification-email-consist-of
	userIDs := core.Map(codes, func(c domain.ActivationCode) uuid.UUID {
		return c.UserID
	})

	const usersQuery = `SELECT * FROM auth.user WHERE id = ANY($1);`
	var users []domain.User
	if err := h.db.SelectContext(ctx, &users, usersQuery, pq.Array(userIDs)); err != nil {
		return core.Unit{}, err
	}

	usersMap := make(map[uuid.UUID]domain.User, len(users))
	for _, user := range users {
		usersMap[user.ID] = user
	}

	var errs []error
	for _, code := range codes {
		email := domain.RegistrationActivationEmail(usersMap[code.UserID], h.emailConfig.Sender)
		if err := h.emailClient.Send(email); err != nil {
			errs = append(errs, err)
		}
	}

	codeIDs := core.Map(codes, func(c domain.ActivationCode) int64 {
		return c.ID
	})

	const updateCodesStmt = `
		UPDATE
			auth.activation_code
		SET
			sent_at = $1
		WHERE
			id = ANY($2);`
	if _, err := h.db.ExecContext(ctx, updateCodesStmt, time.Now().UTC(), pq.Array(codeIDs)); err != nil {
		errs = append(errs, err)
	}

	// Mark all the codes as sent. If the user did not receive the code,
	// they can click the re-send activation code button.
	// It is preferred to not send the email than to send multiple ones.
	return core.Unit{}, errors.Join(errs...)
}
