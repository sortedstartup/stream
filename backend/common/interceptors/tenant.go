package interceptors

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// TenantInterceptor extracts the X-Tenant-ID header from gRPC metadata and puts it in context
func TenantInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Extract metadata from context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			// No metadata, continue without tenant ID
			return handler(ctx, req)
		}

		// Look for X-Tenant-ID header (case-insensitive)
		var tenantID string
		for key, values := range md {
			if strings.ToLower(key) == "x-tenant-id" && len(values) > 0 {
				tenantID = values[0]
				break
			}
		}

		// If tenant ID found, add it to context
		if tenantID != "" {
			ctx = context.WithValue(ctx, "X-Tenant-ID", tenantID)
		}

		return handler(ctx, req)
	}
}
