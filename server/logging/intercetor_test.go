//go:generate go run github.com/golang/mock/mockgen -package=logging -destination=zap_mock_test.go go.uber.org/zap/zapcore ObjectMarshaler

package logging

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jamillosantos/logctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc"
)

func TestInterceptor(t *testing.T) {
	t.Run("should log the operation start and completion", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		wantReqObj := createMockObjectMarshaler(ctrl)
		wantRespObj := createMockObjectMarshaler(ctrl)
		ctx, obs := createObserver()
		called := true
		_, _ = UnaryInterceptor(
			WithOperationStarted(true),
			WithOperationCompleted(true),
			WithRequestExtractor(func(ctx context.Context, req interface{}) (context.Context, zapcore.ObjectMarshaler, error) {
				return ctx, wantReqObj, nil
			}),
			WithResponseExtractor(func(ctx context.Context, resp interface{}) zapcore.ObjectMarshaler {
				return wantRespObj
			}),
		)(ctx, nil, &grpc.UnaryServerInfo{
			Server:     nil,
			FullMethod: "",
		}, func(ctx context.Context, req interface{}) (interface{}, error) {
			called = true
			return map[string]any{}, nil
		})
		require.True(t, called, "the handler should be called")
		entries := obs.All()
		require.Len(t, entries, 2)
		assert.Contains(t, " started", entries[0].Message)
		assert.Len(t, entries[0].Context, 4)
		assert.Equal(t, entries[0].Context[0].Key, fieldGRPCService)
		assert.Equal(t, entries[0].Context[1].Key, fieldGRPCMethod)
		assert.Equal(t, entries[0].Context[2].Key, fieldGRPCFullMethod)
		assert.Equal(t, entries[0].Context[3].Key, fieldGRPCRequest)
		assert.Contains(t, " completed", entries[1].Message)
		assert.Len(t, entries[1].Context, 6)
		assert.Equal(t, entries[1].Context[0].Key, fieldGRPCService)
		assert.Equal(t, entries[1].Context[1].Key, fieldGRPCMethod)
		assert.Equal(t, entries[1].Context[2].Key, fieldGRPCFullMethod)
		assert.Equal(t, entries[1].Context[3].Key, fieldGRPCStatus)
		assert.Equal(t, entries[1].Context[4].Key, fieldGRPCStatusCode)
		assert.Equal(t, entries[1].Context[5].Key, fieldGRPCResponse)
	})

	t.Run("when operation start and completed are disabled", func(t *testing.T) {
		t.Run("should not log the operation start and completion", func(t *testing.T) {
			ctx, obs := createObserver()
			called := true
			_, _ = UnaryInterceptor(
				WithOperationStarted(false),
				WithOperationCompleted(false),
			)(ctx, nil, &grpc.UnaryServerInfo{
				Server:     nil,
				FullMethod: "",
			}, func(ctx context.Context, req interface{}) (interface{}, error) {
				called = true
				return map[string]any{}, nil
			})
			require.True(t, called, "the handler should be called")
			entries := obs.All()
			require.Len(t, entries, 0)
		})

		t.Run("should log completed with error when handler fails", func(t *testing.T) {
			wantErr := errors.New("some error")
			ctx, obs := createObserver()
			_, gotErr := UnaryInterceptor(
				WithOperationStarted(false),
				WithOperationCompleted(false),
				WithRequestExtractor(func(ctx context.Context, req interface{}) (context.Context, zapcore.ObjectMarshaler, error) {
					return ctx, nil, nil
				}),
				WithResponseExtractor(func(ctx context.Context, resp interface{}) zapcore.ObjectMarshaler {
					return nil
				}),
			)(ctx, nil, &grpc.UnaryServerInfo{
				Server:     nil,
				FullMethod: "",
			}, func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, wantErr
			})
			assert.ErrorIsf(t, gotErr, wantErr, "the error should be the same")
			entries := obs.All()
			require.Len(t, entries, 1)

			assert.Contains(t, " completed", entries[0].Message)
			assert.Len(t, entries[0].Context, 5)
			assert.Equal(t, entries[0].Context[0].Key, fieldGRPCService)
			assert.Equal(t, entries[0].Context[1].Key, fieldGRPCMethod)
			assert.Equal(t, entries[0].Context[2].Key, fieldGRPCFullMethod)
			assert.Equal(t, entries[0].Context[3].Key, fieldGRPCStatus)
			assert.Equal(t, entries[0].Context[4].Key, fieldGRPCStatusCode)
		})
	})
}

func createObserver() (context.Context, *observer.ObservedLogs) {
	zc, obs := observer.New(zapcore.DebugLevel)
	return logctx.WithLogger(context.Background(), zap.New(zc)), obs
}

func createMockObjectMarshaler(ctrl *gomock.Controller) zapcore.ObjectMarshaler {
	mock := NewMockObjectMarshaler(ctrl)
	// mock.EXPECT().MarshalLogObject(gomock.Any()).Return(nil)
	return mock
}
