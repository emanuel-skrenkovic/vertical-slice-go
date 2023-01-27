package core

import (
	"encoding/json"
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
		writeBodyIfPresent(w, body)
	}
}

func WriteOK(w http.ResponseWriter, r *http.Request, body interface{}) {
	WriteResponse(w, r, 200, body)
}

func WriteCreated(w http.ResponseWriter, r *http.Request, location string, opts... ResponseOption) {
	// opts = append(opts, WithHeader("Location", location))
	WriteResponse(w, r, 201, nil, WithHeader("Location", location))
}

// TODO: should this accept error as the payload to always convert to a singular respose type?
// (Same for 500 and 502)
func WriteBadRequest(w http.ResponseWriter, r *http.Request, body interface{}) {
	WriteResponse(w, r, 400, body)
}

func WriteInternalServerError(w http.ResponseWriter, r *http.Request, body interface{}) {
	WriteResponse(w, r, 500, body)
}

func WriteBadGateway(w http.ResponseWriter, r *http.Request, body interface{}) {
	WriteResponse(w, r, 502, body)
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
	writeBodyIfPresent(w, body)
}

func writeBodyIfPresent(w http.ResponseWriter, body interface{}) {
	if body == nil {
		return
	}

	// Handle special case where the body is error
	// as error marshals into an empty object.
	if err, ok := body.(error); ok {
		w.Write([]byte(err.Error()))
		return
	}

	responseBytes, err := json.Marshal(body)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	w.Write(responseBytes)
}
