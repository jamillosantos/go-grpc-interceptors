package timeout

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithTimeout(t *testing.T) {
	wantTimeout := time.Second * 123
	opts := defaultOptions()
	WithTimeout(wantTimeout)(&opts)
	assert.Equal(t, opts.defaultTimeout, wantTimeout)
}

func TestTimeout(t *testing.T) {
	t.Run("should apply timeout when configured and given context does not have deadline", func(t *testing.T) {
		wantTimeout := time.Second * 123
		called := false
		_, _ = Timeout(WithTimeout(wantTimeout))(context.Background(), nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			d, ok := ctx.Deadline()
			require.True(t, ok)
			assert.WithinDurationf(t, time.Now().Add(wantTimeout), d, time.Second, "expected deadline to be %s", d)
			called = true
			return nil, nil
		})
		assert.True(t, called, "expected handler to be called")
	})

	t.Run("should not apply any timeout with zero is given", func(t *testing.T) {
		called := false
		_, _ = Timeout(WithTimeout(0))(context.Background(), nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			_, ok := ctx.Deadline()
			assert.False(t, ok)
			called = true
			return nil, nil
		})
		assert.True(t, called, "expected handler to be called")
	})

	t.Run("should not apply any timeout when the given context already has a deadline", func(t *testing.T) {
		wantTimeout := time.Second * 123
		ctx, cancel := context.WithTimeout(context.Background(), wantTimeout)
		defer cancel()
		called := false
		_, _ = Timeout(WithTimeout(0))(ctx, nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			d, ok := ctx.Deadline()
			require.True(t, ok)
			assert.WithinDurationf(t, time.Now().Add(wantTimeout), d, time.Second, "expected deadline to be %s", d)
			called = true
			return nil, nil
		})
		assert.True(t, called, "expected handler to be called")
	})
}
