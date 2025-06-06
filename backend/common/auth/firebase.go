package auth

import (
	"context"
	"encoding/base64"
	"log/slog"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

type Firebase struct {
	App  *firebase.App
	Auth *auth.Client
}

func NewFirebase() (*Firebase, error) {

	//TODO: move it to firebase init
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		//TODO: return right kind of error
		slog.Error("error initializing firebase app", "err", err)
		return nil, err
	}

	auth, err := app.Auth(context.Background())
	if err != nil {
		slog.Error("failed to initialize firebase app using [GOOGLE_APPLICATION_CREDENTIALS]", "err", err)
		slog.Info("trying to initialize firebase app using base64 encoded env variable [GOOGLE_APPLICATION_CREDENTIALS_BASE64]")
		// read creds from env variable
		creds, decodingError := base64.StdEncoding.DecodeString(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_BASE64"))
		if decodingError != nil {
			slog.Error("error decoding base64 firebase credentials from env variable [GOOGLE_APPLICATION_CREDENTIALS_BASE64]", "err", decodingError)
			return nil, decodingError
		}

		app, err = firebase.NewApp(context.Background(), nil, option.WithCredentialsJSON(creds))
		if err != nil {
			slog.Error("error initializing firebase app using base64 encoded env variable [GOOGLE_APPLICATION_CREDENTIALS_BASE64]", "err", err)
			return nil, err
		}

		auth, err = app.Auth(context.Background())
		if err != nil {
			slog.Error("failed to initialize firebase app using base64 encoded env variable [GOOGLE_APPLICATION_CREDENTIALS_BASE64]", "err", err)
			return nil, err
		}

	}

	// return nil
	return &Firebase{
		App:  app,
		Auth: auth,
	}, nil
}

// VerifyIDToken verifies the token and returns the user
func (f *Firebase) VerifyIDToken(token string) (*AuthContext, error) {
	ctx := context.Background()

	// TODO: add docs about when this makes RPC calls
	tok, err := f.Auth.VerifyIDToken(ctx, token)
	if err != nil {
		return &AuthContext{User: &ANONYMOUS, IsAuthenticated: false}, err
	}

	user := &User{
		ID:    tok.UID,
		Name:  tok.Claims["name"].(string),
		Email: tok.Claims["email"].(string),
	}

	return &AuthContext{
		User:            user,
		IsAuthenticated: true,
	}, nil

}
