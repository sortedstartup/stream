package interceptors

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"sortedstartup.com/stream/common/auth"
)

const AUTH_HEADER = "authorization"

type AuthInterceptor func(auth.Auth) grpc.UnaryServerInterceptor

func FirebaseAuthInterceptor(fbauth *auth.Firebase) grpc.UnaryServerInterceptor {

	// The client (browser+JS or language SDK) send the auth token (a string) in the headers of each request
	// We read this auth token from the headers
	// We verify if this auth token in valid using firebase APIs
	// We also decode the token to get the user ID and other details/"claims"
	// We add this user id to the context object
	// This context object is passed to the API handler functions and hence the API know the user ID
	// ---
	// The Auth Token verification is fast
	// -----------------------------------
	// The Auth Token is "digital signed" using a private key and the "signature" verified using a public key by anybody
	// Therefore during verification, we don't need to make any network calls (RPC) to firebase
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// metadata
		authToken, err := getAuthHeader(ctx)
		if err != nil {
			// TODO: This should be printed at trace level since it will be very verbose in logs
			// slog.Debug("No Auth token in request", "err", err)
			return nil, status.Errorf(codes.Unauthenticated, err.Error())
		}

		authContext, verificationErr := fbauth.VerifyIDToken(authToken)

		if verificationErr != nil {
			slog.Info("error verifying ID token", "err", verificationErr)
			return nil, status.Errorf(codes.Unauthenticated, "invalid authentication token")
		}

		newctx := context.WithValue(ctx, auth.AUTH_CONTEXT_KEY, authContext)

		return handler(newctx, req)
	}
}

func FirebaseHTTPAuthMiddleware(fbauth *auth.Firebase, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authorization")

		if authHeader == "" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		authContext, verificationErr := fbauth.VerifyIDToken(authHeader)

		if verificationErr != nil {
			slog.Info("error verifying ID token", "err", verificationErr)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		newctx := context.WithValue(r.Context(), auth.AUTH_CONTEXT_KEY, authContext)
		r = r.WithContext(newctx)
		next.ServeHTTP(w, r)
	})
}

func getAuthHeader(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		// Handle error: metadata is not provided
		return "", fmt.Errorf("missing metadata")
	}

	// Access the Authorization header
	// NOTE: The key is case-insensitive !!
	// 'Authorization' --received-as-> authorization
	authHeaders, ok := md[AUTH_HEADER]

	if !ok || len(authHeaders) == 0 {
		// Handle error: Authorization header is missing
		return "", fmt.Errorf("missing authorization header")
	}

	// Typically, authHeaders is a slice of strings. Use the first value.
	authHeader := authHeaders[0]
	return authHeader, nil
}

// exported function since other packages need this
func AuthFromContext(ctx context.Context) (*auth.AuthContext, error) {
	v, ok := ctx.Value(auth.AUTH_CONTEXT_KEY).(*auth.AuthContext)
	if !ok || v == nil {
		return nil, fmt.Errorf("auth context not found")
	}
	return v, nil
}
