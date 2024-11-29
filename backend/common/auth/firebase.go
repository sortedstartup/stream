package auth

import (
	"context"
	"log/slog"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
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
		slog.Error("error initializing firebase app")
		return nil, err
	}

	auth, err := app.Auth(context.Background())
	if err != nil {
		slog.Error("error initializing firebase auth")
		//TODO: return right kind of error
		return nil, err
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
		return &AuthContext{User: &ANONYMOUS, IsAuthenticated: false, FamilyID: ""}, err
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
