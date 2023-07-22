package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/domain"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"

	"github.com/eskrenkovic/tql"
)

func AuthenticationMiddleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionIDCookie, err := r.Cookie("chess-session")
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

			session, err := tql.QueryFirst[domain.Session](r.Context(), db, q, sessionIDCookie.Value)
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

			authContext := context.WithValue(r.Context(), core.SessionContextKey, session)
			next.ServeHTTP(w, r.WithContext(authContext))
		})
	}
}
