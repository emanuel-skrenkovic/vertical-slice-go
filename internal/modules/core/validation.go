package core

import (
	"context"
	"strings"

	"github.com/eskrenkovic/mediator-go"
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

type RequestValidationBehavior struct {}

func (b *RequestValidationBehavior) Handle(
	ctx context.Context,
	request interface{},
	next mediator.RequestHandlerFunc,
) (interface{}, error) {
	if request, ok := request.(Validator); ok {
		if err := request.Validate(); err != nil {
			return nil, NewCommandError(400, err, "request validation failed")
		}
	}

	return next(ctx, request)
}
