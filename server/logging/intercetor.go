package logging

import (
	"context"
	"fmt"
	"strings"

	"github.com/jamillosantos/logctx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type contextAddFieldUpdater string

func (s contextAddFieldUpdater) String() string {
	return string(s) + "contextAddFieldUpdater"
}

const (
	fieldGRPCService      = "grpc.service"
	fieldGRPCMethod       = "grpc.method"
	fieldGRPCFullMethod   = "grpc.full_method"
	fieldGRPCStatus       = "grpc.status"
	fieldGRPCStatusCode   = "grpc.status_code"
	fieldGRPCRequest      = "grpc.request"
	fieldGRPCResponse     = "grpc.response"
	fieldGRPCErrorMessage = "grpc.error.message"
	fieldGRPCErrorDetails = "grpc.error.details"
)

const (
	messageRequest       = "%s started"
	messageResponse      = "%s completed"
	messageResponseError = "%s completed with error"
)

type loggingOptions struct {
	extractRequest       func(ctx context.Context, req interface{}) (context.Context, zapcore.ObjectMarshaler, error)
	extractResponse      func(ctx context.Context, resp interface{}) zapcore.ObjectMarshaler
	handleError          func(ctx context.Context, err error) []zap.Field
	logRequest           bool
	requestMessage       string
	logResponse          bool
	responseMessage      string
	responseErrorMessage string
}

type Option func(*loggingOptions)

func defaultOptions() loggingOptions {
	return loggingOptions{
		extractRequest:       nil,
		extractResponse:      nil,
		handleError:          defaultHandleError,
		logRequest:           false,
		requestMessage:       messageRequest,
		logResponse:          true,
		responseMessage:      messageResponse,
		responseErrorMessage: messageResponseError,
	}
}

func UnaryInterceptor(options ...Option) grpc.UnaryServerInterceptor {
	opts := defaultOptions()
	for _, opt := range options {
		opt(&opts)
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var (
			reqObj zapcore.ObjectMarshaler
		)
		if opts.extractRequest != nil {
			c, reqZapObj, err := opts.extractRequest(ctx, req)
			if err != nil {
				return nil, err
			}
			reqObj = reqZapObj
			if c != nil {
				ctx = c
			}
		}
		service, method := extractServiceAndMethod(info.FullMethod)
		commonFields := buildCommonFields(service, method, info)

		ctx = logRequest(ctx, method, commonFields, reqObj, opts)
		resp, err = handler(ctx, req)

		var respObj zapcore.ObjectMarshaler
		if opts.extractResponse != nil {
			respObj = opts.extractResponse(ctx, resp)
		}
		logResponse(ctx, method, commonFields, reqObj, respObj, err, opts)
		return
	}
}

func extractServiceAndMethod(fullMethod string) (string, string) {
	toks := strings.Split(fullMethod[strings.LastIndexByte(fullMethod, '.')+1:], "/")
	if len(toks) == 2 {
		return toks[0], toks[1]
	}
	return "", ""
}

func logRequest(ctx context.Context, method string, fields []zap.Field, reqObj zapcore.ObjectMarshaler, opts loggingOptions) context.Context {
	if !opts.logRequest {
		return ctx
	}

	if reqObj != nil {
		fields = append(fields, zap.Object(fieldGRPCRequest, reqObj))
	}

	logctx.Info(ctx, fmt.Sprintf(opts.requestMessage, method), fields...)
	return ctx
}

func logResponse(ctx context.Context, method string, fields []zap.Field, reqObj zapcore.ObjectMarshaler, respObj zapcore.ObjectMarshaler, err error, opts loggingOptions) {
	if !opts.logResponse && err == nil {
		return
	}

	stCode := codes.OK
	if s, ok := status.FromError(err); ok {
		stCode = s.Code()
	}
	fields = append(
		fields,
		zap.String(fieldGRPCStatus, stCode.String()),
		zap.Uint32(fieldGRPCStatusCode, uint32(stCode)),
	)

	if !opts.logRequest && reqObj != nil {
		fields = append(fields, zap.Object(fieldGRPCRequest, reqObj))
	}

	if respObj != nil {
		fields = append(fields, zap.Object(fieldGRPCResponse, respObj))
	}

	writeLog := logctx.Info
	logMessage := opts.responseMessage

	if err != nil {
		writeLog = logctx.Error
		logMessage = opts.responseErrorMessage
		if opts.handleError != nil {
			fields = append(fields, opts.handleError(ctx, err)...)
		}
	}

	writeLog(ctx, fmt.Sprintf(logMessage, method), fields...)
}

func buildCommonFields(service string, method string, info *grpc.UnaryServerInfo) []zap.Field {
	f := make([]zap.Field, 0, 4)
	f = append(f, zap.String(fieldGRPCService, service)) // TODO Extract service name
	f = append(f, zap.String(fieldGRPCMethod, method))
	return append(f, zap.String(fieldGRPCFullMethod, info.FullMethod))
}
