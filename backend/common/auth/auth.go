package auth

type User struct {
	ID    string
	Name  string
	Email string
	Roles []Role
}

var ANONYMOUS = User{
	ID:    "anonymous",
	Name:  "anonymous",
	Email: "anonymous@example.com",
	Roles: []Role{},
}

type Role string

// some hardcoded roles
const (
	Admin Role = "admin"
)

type AuthContextKey string

const AUTH_CONTEXT_KEY AuthContextKey = "auth"

type AuthContext struct {
	User            *User
	FamilyID        string
	IsAuthenticated bool
}

var ANONYMOUS_AUTH_CTX AuthContext = AuthContext{User: &ANONYMOUS, IsAuthenticated: false, FamilyID: ""}

type Auth interface {
	VerifyIDToken(token string) (authCtx *AuthContext, err error)
}
