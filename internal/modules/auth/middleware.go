package auth

import (
	"context"
	"database/sql"
	"errors"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/domain"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"
	"github.com/jmoiron/sqlx"
	"net/http"
)

type authContextKey string

const sessionContextKey authContextKey = "session"

func AuthenticationMiddleware(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionID, err := r.Cookie("chess-session")
			if err != nil {
				core.WriteUnauthorized(w, r, nil)
				return
			}

			const q = `
				SELECT
				    *
				FROM
				    auth.session
				WHERE
					id = $1;`

			var session domain.Session
			err = db.GetContext(r.Context(), &session, q, sessionID)
			switch {
			case err != nil && errors.Is(err, sql.ErrNoRows):
				core.WriteUnauthorized(w, r, nil)
				return
			case err != nil:
				core.WriteInternalServerError(w, r, nil)
				return
			}

			if err := session.Validate(); err != nil {
				core.WriteUnauthorized(w, r, nil)
				return
			}

			authContext := context.WithValue(r.Context(), sessionContextKey, session)
			next.ServeHTTP(w, r.WithContext(authContext))
		})
	}
}
