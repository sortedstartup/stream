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

// Shared Firebase token verification logic
func verifyFirebaseToken(fbauth *auth.Firebase, token string) (*auth.AuthContext, error) {
	authContext, verificationErr := fbauth.VerifyIDToken(token)
	if verificationErr != nil {
		slog.Info("error verifying ID token", "err", verificationErr)
		return nil, fmt.Errorf("invalid authentication token")
	}
	return authContext, nil
}

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

		authContext, verificationErr := verifyFirebaseToken(fbauth, authToken)
		if verificationErr != nil {
			return nil, status.Errorf(codes.Unauthenticated, verificationErr.Error())
		}

		newctx := context.WithValue(ctx, auth.AUTH_CONTEXT_KEY, authContext)

		return handler(newctx, req)
	}
}

func FirebaseHTTPHeaderAuthMiddleware(fbauth *auth.Firebase, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("authorization")

		if authHeader == "" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		authContext, verificationErr := verifyFirebaseToken(fbauth, authHeader)
		if verificationErr != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// After successful authorization header verification, manage the authorization cookie
		const authCookieName = "authorization"

		// Check if existing cookie is present
		existingCookie, err := r.Cookie(authCookieName)

		shouldSetCookie := false

		if err != nil {
			// Cookie doesn't exist, set it
			shouldSetCookie = true
			slog.Debug("Authorization cookie not present, will set it")
		} else if existingCookie.Value != authHeader {
			// Cookie exists but has different value, update it
			shouldSetCookie = true
			slog.Debug("Authorization cookie has different value, will update it")
		}

		if shouldSetCookie {
			authCookie := &http.Cookie{
				Name:     authCookieName,
				Value:    authHeader,
				Path:     "/",
				MaxAge:   3600 * 24, // 1 day
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
			}
			http.SetCookie(w, authCookie)
			slog.Debug("Authorization cookie set/updated")
		}

		newctx := context.WithValue(r.Context(), auth.AUTH_CONTEXT_KEY, authContext)
		r = r.WithContext(newctx)
		next.ServeHTTP(w, r)
	})
}

func FirebaseCookieAuthMiddleware(fbauth *auth.Firebase, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const authCookieName = "authorization"

		// Get authorization cookie
		authCookie, err := r.Cookie(authCookieName)
		if err != nil {
			slog.Debug("Authorization cookie not found", "err", err, "url", r.URL.Path)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if authCookie.Value == "" {
			slog.Debug("Authorization cookie is empty", "url", r.URL.Path)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Verify the token from cookie using shared logic
		authContext, verificationErr := verifyFirebaseToken(fbauth, authCookie.Value)
		if verificationErr != nil {
			slog.Info("Invalid authorization cookie", "err", verificationErr, "url", r.URL.Path)
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
