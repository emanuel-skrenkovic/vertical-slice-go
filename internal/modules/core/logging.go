package core

import (
	"context"
	"log/slog"

	"github.com/eskrenkovic/mediator-go"
)

type loggingContextKey string

const contextLoggerKey loggingContextKey = "context_logger"

func LogError(ctx context.Context, message string, fields ...any) {
	logger := ctx.Value(contextLoggerKey)
	slogLogger, success := logger.(*slog.Logger)
	if !success {
		panic("failed to convert context logger to *slog.Logger")
	}

	slogLogger.Error(message, fields...)
}

var _ mediator.PipelineBehavior = (*RequestLoggingBehavior)(nil)

type RequestLoggingBehavior struct {
	Logger *slog.Logger
}

func (b *RequestLoggingBehavior) Handle(
	ctx context.Context,
	request any,
	next mediator.RequestHandlerFunc,
) (any, error) {
	var logFields []any

	requestID := ctx.Value(0)
	if requestID != nil && requestID != "" {
		logFields = append(logFields, []any{"request_id", requestID}...)
	}

	correlationID := ctx.Value(CorrelationIDContextKey)
	if correlationID != nil && correlationID != "" {
		logFields = append(logFields, []any{"correlation_id", correlationID}...)
	}

	if request != nil {
		logFields = append(logFields, []any{"request_body", request}...)
	}

	b.Logger.Info("processing request", logFields...)

	return next(ctx, request)
}

var _ mediator.PipelineBehavior = (*HandlerErrorLoggingBehavior)(nil)

type HandlerErrorLoggingBehavior struct {
	Logger *slog.Logger
}

func (b *HandlerErrorLoggingBehavior) Handle(
	ctx context.Context,
	request any,
	next mediator.RequestHandlerFunc,
) (any, error) {
	response, err := next(ctx, request)
	if err != nil {
		b.Logger.Error("handler returned error", "error", err)
	}

	return response, err
}
