package core

import "fmt"

type Unit struct{}

type CommandError struct {
	Payload    interface{}
	StatusCode int
	Reason     *string
}

type CommandErrorOption func(*CommandError)

func WithReason(reason string) CommandErrorOption {
	return func(e *CommandError) {
		e.Reason = &reason
	}
}

func NewCommandError(statusCode int, payload interface{}, opts ...CommandErrorOption) CommandError {
	e := CommandError{
		StatusCode: statusCode,
		Payload:    payload,
	}

	for _, opt := range opts {
		opt(&e)
	}

	return e
}

func (r CommandError) Error() string {
	var values struct {
		Payload    interface{}
		StatusCode int
		Reason     string
	}

	values.Payload = r.Payload
	values.StatusCode = r.StatusCode

	if r.Reason != nil {
		values.Reason = *r.Reason
	}

	return fmt.Sprintf("%+v", values)
}
