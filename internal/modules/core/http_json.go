package core

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
)

func RequestBody[TRequest any](r *http.Request) (TRequest, error) {
	var request TRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	return request, err
}

type ResponseOption func(http.ResponseWriter, *http.Request)

func WithHeader(header, value string) ResponseOption {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(header, value)
	}
}

func WithBody(body interface{}) ResponseOption {
	return func(w http.ResponseWriter, r *http.Request) {
		writeBodyIfPresent(r.Context(), w, body)
	}
}

func WriteOK(w http.ResponseWriter, r *http.Request, body interface{}) {
	WriteResponse(w, r, 200, body)
}

func WriteCreated(w http.ResponseWriter, r *http.Request, location string, opts ...ResponseOption) {
	// opts = append(opts, WithHeader("Location", location))
	WriteResponse(w, r, 201, nil, WithHeader("Location", location))
}

func WriteBadRequest(w http.ResponseWriter, r *http.Request, body interface{}) {
	WriteResponse(w, r, 400, body)
}

func WriteUnauthorized(w http.ResponseWriter, r *http.Request, body interface{}) {
	WriteResponse(w, r, 401, body)
}

func WriteInternalServerError(w http.ResponseWriter, r *http.Request, body interface{}) {
	WriteResponse(w, r, 500, body)
}

func WriteBadGateway(w http.ResponseWriter, r *http.Request, body interface{}) {
	WriteResponse(w, r, 502, body)
}

func WriteCommandError(w http.ResponseWriter, r *http.Request, err error, opts ...ResponseOption) {
	statusCode := 500
	if commandErr, ok := err.(CommandError); ok {
		statusCode = commandErr.StatusCode
	}
	WriteResponse(w, r, statusCode, err)
}

func WriteResponse(
	w http.ResponseWriter,
	r *http.Request,
	statusCode int,
	body interface{},
	opts ...ResponseOption,
) {
	for _, opt := range opts {
		opt(w, r)
	}
	w.WriteHeader(statusCode)
	writeBodyIfPresent(r.Context(), w, body)
}

func writeBodyIfPresent(ctx context.Context, w http.ResponseWriter, body interface{}) {
	if body == nil {
		return
	}

	// Handle special case where the body is error
	// as error marshals into an empty object.
	if err, ok := body.(error); ok {
		responseBytes, err := json.Marshal(err)
		if err != nil {
			LogError(ctx, "failed to serialize response error", zap.Error(err))
			return
		}

		if _, err := w.Write(responseBytes); err != nil {
			LogError(ctx, "failed to write response", zap.Error(err))
		}
		return
	}

	responseBytes, err := json.Marshal(body)
	if err != nil {
		if _, err := w.Write([]byte(err.Error())); err != nil {
			LogError(ctx, "failed to write response", zap.Error(err))
		}
		return
	}

	if _, err := w.Write(responseBytes); err != nil {
		LogError(ctx, "failed to write response", zap.Error(err))
	}
}
