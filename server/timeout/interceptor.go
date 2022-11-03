package timeout

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

type opts struct {
	defaultTimeout time.Duration
}

// Option is a function that configures the Timeout interceptor.
type Option func(*opts)

func defaultOptions() opts {
	return opts{
		defaultTimeout: time.Second * 10,
	}
}

// Timeout is the interceptor that will add a timeout to the context of the request if it is not already set.
// The default timeout added is 10s. You can customize it by specifying WihtTimeout option.
func Timeout(opts ...Option) grpc.UnaryServerInterceptor {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if o.defaultTimeout == 0 {
			return handler(ctx, req)
		}
		// If a deadline isn't defined yet...
		if _, ok := ctx.Deadline(); !ok {
			c, cancelFnc := context.WithTimeout(ctx, o.defaultTimeout)
			defer cancelFnc()
			ctx = c
		}
		return handler(ctx, req)
	}
}

// WithTimeout is an Option sets the timeout for the interceptor.
func WithTimeout(timeout time.Duration) Option {
	return func(o *opts) {
		o.defaultTimeout = timeout
	}
}
