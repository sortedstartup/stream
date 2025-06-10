package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

type Firebase struct {
	App  *firebase.App
	Auth *auth.Client
}

func NewFirebase() (*Firebase, error) {
	// Try default credentials first
	app, auth, err := initializeFirebaseApp()
	if err == nil {
		return &Firebase{App: app, Auth: auth}, nil
	}

	slog.Error("failed to initialize firebase app using default credentials", "err", err)

	// Fallback to base64 encoded credentials
	app, auth, err = initializeFirebaseWithBase64Creds()
	if err != nil {
		slog.Error("failed to initialize firebase app using base64 credentials", "err", err)
		return nil, err
	}

	return &Firebase{App: app, Auth: auth}, nil
}

// initializeFirebaseApp initializes Firebase using default credentials
func initializeFirebaseApp() (*firebase.App, *auth.Client, error) {
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		return nil, nil, err
	}

	authClient, err := app.Auth(context.Background())
	if err != nil {
		return nil, nil, err
	}

	return app, authClient, nil
}

// initializeFirebaseWithBase64Creds initializes Firebase using base64 encoded credentials
func initializeFirebaseWithBase64Creds() (*firebase.App, *auth.Client, error) {
	slog.Info("trying to initialize firebase app using base64 encoded env variable [GOOGLE_APPLICATION_CREDENTIALS_BASE64]")

	creds, err := base64.StdEncoding.DecodeString(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_BASE64"))
	if err != nil {
		return nil, nil, err
	}

	app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsJSON(creds))
	if err != nil {
		return nil, nil, err
	}

	authClient, err := app.Auth(context.Background())
	if err != nil {
		return nil, nil, err
	}

	return app, authClient, nil
}

// isEmailAllowed checks if the email is in the allowed list
func isEmailAllowed(email string) bool {
	allowedEmails := os.Getenv("ALLOWED_EMAILS")
	if allowedEmails == "" {
		// If no whitelist is set, allow all emails (default behavior)
		return true
	}

	// Split the comma-separated list and check if email is in the list
	emailList := strings.Split(allowedEmails, ",")
	for _, allowedEmail := range emailList {
		if strings.TrimSpace(allowedEmail) == email {
			return true
		}
	}
	return false
}

// VerifyIDToken verifies the token and returns the user
func (f *Firebase) VerifyIDToken(token string) (*AuthContext, error) {
	ctx := context.Background()

	// TODO: add docs about when this makes RPC calls
	tok, err := f.Auth.VerifyIDToken(ctx, token)
	if err != nil {
		return &AuthContext{User: &ANONYMOUS, IsAuthenticated: false}, err
	}

	email := tok.Claims["email"].(string)

	// Check if email is in the allowed list
	if !isEmailAllowed(email) {
		return &AuthContext{User: &ANONYMOUS, IsAuthenticated: false}, fmt.Errorf("email not in allowed list: %s", email)
	}

	user := &User{
		ID:    tok.UID,
		Name:  tok.Claims["name"].(string),
		Email: email,
	}

	return &AuthContext{
		User:            user,
		IsAuthenticated: true,
	}, nil
}
