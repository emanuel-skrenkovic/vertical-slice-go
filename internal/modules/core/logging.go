package core

import (
	"context"

	"github.com/eskrenkovic/mediator-go"

	"go.uber.org/zap"
)

var _ mediator.PipelineBehavior = (*RequestLoggingBehavior)(nil)

type RequestLoggingBehavior struct {
	Logger *zap.Logger
}

func (b *RequestLoggingBehavior) Handle(
	ctx context.Context,
	request interface{},
	next mediator.RequestHandlerFunc,
) (interface{}, error) {
	if request != nil {
		b.Logger.Info("processing request", zap.Any("request", request))
	}

	return next(ctx, request)
}

var _ mediator.PipelineBehavior = (*HandlerErrorLoggingBehavior)(nil)

type HandlerErrorLoggingBehavior struct {
	Logger *zap.Logger
}

func (b *HandlerErrorLoggingBehavior) Handle(
	ctx context.Context,
	request interface{},
	next mediator.RequestHandlerFunc,
) (interface{}, error) {
	response, err := next(ctx, request)
	if err != nil {
		b.Logger.Error("handler returned error", zap.Error(err))
	}

	return response, err
}
