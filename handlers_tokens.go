package main

import (
	"fmt"
	"net/http"

	"github.com/seiobata/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefreshAccess(w http.ResponseWriter, r *http.Request) {
	type response struct {
		AccessToken string `json:"token"`
	}

	// verify refresh token
	invalidErr := "Token is invalid or expired"
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		helperResponseError(w, http.StatusUnauthorized, invalidErr)
		return
	}
	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		helperResponseError(w, http.StatusUnauthorized, invalidErr)
		return
	}

	// generate new access token
	accessToken, err := auth.MakeJWT(user.ID, cfg.secret, accessTkExp)
	if err != nil {
		makeJWTErr := fmt.Sprintf("Error making JWT token: %v", err)
		helperResponseError(w, http.StatusInternalServerError, makeJWTErr)
		return
	}

	helperResponseJSON(w, http.StatusOK, response{
		AccessToken: accessToken,
	})
}
