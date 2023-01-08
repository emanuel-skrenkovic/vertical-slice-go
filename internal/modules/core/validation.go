package core

import (
	"strings"
)

type Validator interface {
	Validate() error
}

type ValidationError struct {
	ValidationErrors []error
}

func (e ValidationError) Error() string {
	var b strings.Builder
	for _, err := range e.ValidationErrors {
		b.WriteString(" '")
		b.WriteString(err.Error())
		b.WriteString("'")
	}
	return b.String()
}
