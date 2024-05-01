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

type RequestValidationBehavior struct{}

func (b *RequestValidationBehavior) Handle(
	ctx context.Context,
	request any,
	next mediator.RequestHandlerFunc,
) (any, error) {
	v, ok := request.(Validator)
	if !ok {
		return next(ctx, request)
	}

	if err := v.Validate(); err != nil {
		return nil, NewCommandError(400, err, WithReason("request validation failed"))
	}

	return next(ctx, request)
}
