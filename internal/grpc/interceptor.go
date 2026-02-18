package grpc

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type contextKey string

const requestIDKey contextKey = "requestID"

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func RequestIDInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		requestID := uuid.NewString()
		ctx = context.WithValue(ctx, requestIDKey, requestID)

		return handler(ctx, req)
	}
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

func StreamRequestIDInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	requestID := uuid.NewString()

	ctx := context.WithValue(ss.Context(), requestIDKey, requestID)
	return handler(srv, &wrappedStream{ServerStream: ss, ctx: ctx})
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(requestIDKey).(string)
	return requestID, ok
}
