package logging

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestWithRequestStart(t *testing.T) {
	var opts loggingOptions
	WithOperationStarted(true)(&opts)
	assert.True(t, opts.logRequest)
}

func TestWithRequestEnd(t *testing.T) {
	var opts loggingOptions
	WithOperationCompleted(true)(&opts)
	assert.True(t, opts.logResponse)
}

func TestWithRequestAndContextFieldsExtractor(t *testing.T) {
	var opts loggingOptions
	WithRequestExtractor(func(ctx context.Context, req interface{}) (context.Context, zapcore.ObjectMarshaler, error) {
		return ctx, nil, nil
	})(&opts)
	assert.NotNil(t, opts.extractRequest)
}

func TestWithResponseFieldsExtractor(t *testing.T) {
	var opts loggingOptions
	WithResponseExtractor(func(ctx context.Context, resp interface{}) zapcore.ObjectMarshaler { return nil })(&opts)
	assert.NotNil(t, opts.extractResponse)
}

func TestWithErrorHandler(t *testing.T) {
	var opts loggingOptions
	WithErrorHandler(func(ctx context.Context, err error) []zapcore.Field { return nil })(&opts)
	assert.NotNil(t, opts.handleError)
}
