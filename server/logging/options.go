package logging

import (
	"context"

	"go.uber.org/zap/zapcore"
)

func WithOperationStarted(enable bool) Option {
	return func(opts *loggingOptions) {
		opts.logRequest = enable
	}
}

func WithOperationCompleted(enable bool) Option {
	return func(opts *loggingOptions) {
		opts.logResponse = enable
	}
}

func WithRequestExtractor(extractor func(ctx context.Context, req interface{}) (context.Context, zapcore.ObjectMarshaler, error)) Option {
	return func(opts *loggingOptions) {
		opts.extractRequest = extractor
	}
}

func WithResponseExtractor(extractor func(ctx context.Context, resp interface{}) zapcore.ObjectMarshaler) Option {
	return func(opts *loggingOptions) {
		opts.extractResponse = extractor
	}
}

func WithErrorHandler(handler func(ctx context.Context, err error) []zapcore.Field) Option {
	return func(opts *loggingOptions) {
		opts.handleError = handler
	}
}
