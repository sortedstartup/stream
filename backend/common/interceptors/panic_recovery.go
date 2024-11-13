package interceptors

import (
	"context"
	"log"
	"runtime"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// panicRecoveryInterceptor returns a new unary server interceptor that recovers from panics.
func PanicRecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if p := recover(); p != nil {
				// Capture the stack trace
				panicRecoveryHandler(p)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

// panicRecoveryHandler handles the panic for both unary and stream interceptors.
func panicRecoveryHandler(p interface{}) error {
	stack := make([]byte, 4096)
	length := runtime.Stack(stack, false)
	log.Printf("recovered from a panic: %v\n%s", p, stack[:length])

	return status.Errorf(codes.Internal, "internal server error")
}
