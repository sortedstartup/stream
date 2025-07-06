package interceptors

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const TENANT_ID_HEADER = "x-tenant-id"

// Define a custom type for context keys to avoid collisions
type contextKey string

const tenantIDKey contextKey = "tenant-id"

// TenantInterceptor extracts the x-tenant-id header from gRPC metadata and puts it in context
func TenantInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Extract metadata from context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			// Handle error: metadata is not provided
			return "", fmt.Errorf("missing metadata")
		}

		// Look for x-tenant-id header (case-insensitive)
		// NOTE: The key is case-insensitive !!
		// 'x-tenant-id' --received-as-> x-tenant-id
		tenantIDHeaders, ok := md[TENANT_ID_HEADER]

		// we did not find the tenant ID header
		if !ok || len(tenantIDHeaders) == 0 {
			// No tenant ID found, continue without tenant ID
			return handler(ctx, req)
		}

		tenantID := tenantIDHeaders[0]
		// If tenant ID found, add it to context
		if tenantID != "" {
			ctx = context.WithValue(ctx, tenantIDKey, tenantID)
		}

		// Typically, tenantID is a slice of strings. Use the first value.
		return handler(ctx, req)
	}
}

// GetTenantIDFromContext retrieves the tenant ID from the context
func GetTenantIDFromContext(ctx context.Context) (string, error) {
	tenantID, ok := ctx.Value(tenantIDKey).(string)
	if !ok || tenantID == "" {
		return "", fmt.Errorf("tenant ID not found in context")
	}
	return tenantID, nil
}
