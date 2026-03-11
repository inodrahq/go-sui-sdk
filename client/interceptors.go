package client

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// applyTimeout adds a timeout to ctx if timeout > 0 and the context
// does not already have an earlier deadline.
func applyTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return ctx, func() {}
	}
	if deadline, ok := ctx.Deadline(); ok && time.Until(deadline) <= timeout {
		return ctx, func() {} // caller's deadline is tighter
	}
	return context.WithTimeout(ctx, timeout)
}

// headerInterceptor returns a unary interceptor that injects metadata headers
// and applies a default timeout.
func headerInterceptor(headers map[string]string, timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx, cancel := applyTimeout(ctx, timeout)
		defer cancel()
		if len(headers) > 0 {
			md := metadata.New(headers)
			ctx = metadata.NewOutgoingContext(ctx, md)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// headerStreamInterceptor returns a stream interceptor that injects metadata headers
// and applies a default timeout.
func headerStreamInterceptor(headers map[string]string, timeout time.Duration) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx, cancel := applyTimeout(ctx, timeout)
		defer cancel()
		if len(headers) > 0 {
			md := metadata.New(headers)
			ctx = metadata.NewOutgoingContext(ctx, md)
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}
