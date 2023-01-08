package core

import "fmt"

type Unit struct{}

type CommandError struct {
	Payload    interface{}
	StatusCode int
	Reason     *string
}

func NewCommandError(statusCode int, payload interface{}, reason string) CommandError {
	return CommandError{
		StatusCode: statusCode,
		Payload:    payload,
		Reason:     &reason,
	}
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
