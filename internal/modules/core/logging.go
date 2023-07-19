package core

import (
	"context"

	"github.com/eskrenkovic/mediator-go"

	"go.uber.org/zap"
)

type loggingContextKey string

const contextLoggerKey loggingContextKey = "context_logger"

func LogError(ctx context.Context, message string, fields ...zap.Field) {
	logger := ctx.Value(contextLoggerKey)
	zapLogger, success := logger.(*zap.Logger)
	if !success {
		panic("failed to convert context logger to *zap.Logger")
	}

	zapLogger.Error(message, fields...)
}

var _ mediator.PipelineBehavior = (*RequestLoggingBehavior)(nil)

type RequestLoggingBehavior struct {
	Logger *zap.Logger
}

func (b *RequestLoggingBehavior) Handle(
	ctx context.Context,
	request interface{},
	next mediator.RequestHandlerFunc,
) (interface{}, error) {
	var logFields []zap.Field

	requestID := ctx.Value(0)
	if requestID != nil && requestID != "" {
		logFields = append(logFields, zap.Any("request_id", requestID))
	}

	correlationID := ctx.Value(CorrelationIDContextKey)
	if correlationID != nil && correlationID != "" {
		logFields = append(logFields, zap.Any("correlation_id", correlationID))
	}

	if request != nil {
		logFields = append(logFields, zap.Any("request_body", request))
	}

	b.Logger.Info("processing request", logFields...)

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
