package grpc

import (
	"chronos-queue/internal/requestid"
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

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
		ctx = requestid.WithRequestID(ctx, requestID)

		return handler(ctx, req)
	}
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

func StreamRequestIDInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	requestID := uuid.NewString()
	ctx := requestid.WithRequestID(ss.Context(), requestID)

	return handler(srv, &wrappedStream{ServerStream: ss, ctx: ctx})
}
