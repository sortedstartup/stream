package api

import (
	"net/http"

	"sortedstartup.com/stream/common/interceptors"
)

func (api *UserAPI) CreateUserIfNotExistsHandler(w http.ResponseWriter, r *http.Request) {
	api.log.Info("CreateUserIfNotExistsHandler")

	authContext, err := interceptors.AuthFromContext(r.Context())
	if err != nil {
		api.log.Error("Auth context not found in request", "err", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := authContext.User.ID

	user, err := api.dbQueries.GetUserByID(r.Context(), userID)
	if err != nil {
		api.log.Error("Error getting user", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		api.log.Info("User not found, creating user")

	}
}
