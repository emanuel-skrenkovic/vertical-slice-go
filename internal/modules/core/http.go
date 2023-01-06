package core

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const (
	CorrelationIDHeader     = "Correlation-Id"
	CorrelationIDContextKey = "correlation_id"
)

func CorrelationIDHTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		correlationID := r.Header.Get(CorrelationIDHeader)
		if correlationID == "" {
			correlationID = uuid.NewString()
		}

		ctx = context.WithValue(ctx, CorrelationIDContextKey, correlationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
