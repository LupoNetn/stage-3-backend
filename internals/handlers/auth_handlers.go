package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/luponetn/hng-stage-1/internals/db"
	httprequest "github.com/luponetn/hng-stage-1/internals/httpRequest"
)

func (h *Handler) HandleGithubCLIAuth(w http.ResponseWriter, r *http.Request) {
	var req GithubCLIAuth
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Println("Error decoding request:", err)
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]any{
			"status": "error",
			"message": "invalid request",
		})
		return
	}

    fmt.Println(req.Code, req.CodeVerifier, req.State)
	githubTokenUrl := os.Getenv("GITHUB_OAUTH_TOKEN_URL")
	githubClientId := os.Getenv("CLI_GITHUB_CLIENT_ID")
	githubClientSecret := os.Getenv("CLI_GITHUB_CLIENT_SECRET")

	body := struct {
		Code         string `json:"code"`
		CodeVerifier string `json:"code_verifier"`
		State        string `json:"state"`
		ClientId     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}{
		Code:         req.Code,
		CodeVerifier: req.CodeVerifier,
		State:        req.State,
		ClientId:     githubClientId,
		ClientSecret: githubClientSecret,
	}

	data, err := httprequest.MakeRequest(r.Context(), "POST", githubTokenUrl, body)
	if err != nil {
		fmt.Println("Error exchanging code for token:", err)
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]any{
			"status": "error",
			"message": "failed to exchange code for token",
		})
		return
	}
	// Phase 5: Retrieve GitHub User
	tokenMap, ok := data.(map[string]any)
	if !ok {
		fmt.Println("Error parsing token response")
		w.WriteHeader(500)
		return
	}
	
	accessToken, ok := tokenMap["access_token"].(string)
	if !ok {
		fmt.Println("Access token not found in response")
		w.WriteHeader(500)
		return
	}

	githubUserUrl := os.Getenv("GITHUB_USER_URL")
	if githubUserUrl == "" {
		githubUserUrl = "https://api.github.com/user"
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", accessToken),
	}

	userData, err := httprequest.MakeRequestWithHeaders(r.Context(), "GET", githubUserUrl, nil, headers)
	if err != nil {
		fmt.Println("Error fetching github user:", err)
		w.WriteHeader(500)
		return
	}

	userMap, ok := userData.(map[string]any)
	if ok {
		fmt.Printf("login: %v\nid: %v\n", userMap["login"], userMap["id"])

		// User Creation / Login
		githubIDStr := fmt.Sprintf("%v", userMap["id"])
		usernameStr := fmt.Sprintf("%v", userMap["login"])
		emailStr := ""
		if email, ok := userMap["email"].(string); ok {
			emailStr = email
		}
		avatarUrlStr := ""
		if avatar, ok := userMap["avatar_url"].(string); ok {
			avatarUrlStr = avatar
		}

		var finalUser db.User
		existingUser, err := h.queries.GetUserByGithubID(r.Context(), githubIDStr)
		if err == nil {
			// User exists -> login
			h.queries.UpdateLastLogin(r.Context(), existingUser.ID)
			finalUser = existingUser
		} else {
			// User doesn't exist -> create user
			newID, _ := uuid.NewV7()
			finalUser, err = h.queries.CreateUser(r.Context(), db.CreateUserParams{
				ID:          toUUID(newID),
				GithubID:    githubIDStr,
				Username:    usernameStr,
				Email:       emailStr,
				AvatarUrl:   toText(avatarUrlStr),
				Role:        "analyst",
				IsActive:    true,
				LastLoginAt: toTimestamptz(time.Now()),
			})
			if err != nil {
				fmt.Println("Error creating user:", err)
				w.WriteHeader(500)
				return
			}
		}

		// Phase 7: Generate Tokens
		accessToken, _ := generateToken(finalUser.ID, finalUser.Username, 3*time.Minute)
		refreshToken, _ := generateToken(finalUser.ID, finalUser.Username, 5*time.Minute)

		// Store refresh token
		h.queries.UpdateRefreshToken(r.Context(), db.UpdateRefreshTokenParams{
			ID:           finalUser.ID,
			RefreshToken: toText(refreshToken),
		})

		// Phase 8: Return Tokens to CLI
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"username":      finalUser.Username,
		})
		return
	}

	h.errorResponse(w, http.StatusInternalServerError, "failed to retrieve user profile from GitHub")
}

func generateToken(userID pgtype.UUID, username string, duration time.Duration) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "super-secret-key-for-dev"
	}

	claims := jwt.MapClaims{
		"user_id":  uuid.UUID(userID.Bytes).String(),
		"username": username,
		"exp":      time.Now().Add(duration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}