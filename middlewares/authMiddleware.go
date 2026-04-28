package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/luponetn/hng-stage-1/utils"
)


type contextKey string

const UserContextKey contextKey = "user"

func GetUserClaims(ctx context.Context) (*utils.Claims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*utils.Claims)
	return claims, ok
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string

		cookie, err := r.Cookie("access_token")
		if err == nil {
			tokenString = cookie.Value
		} else {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if tokenString == "" {
			ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		claims, err := utils.VerifyToken(tokenString)
		if err != nil {
			ErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		// Correctly create and pass the new context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AuthorizeAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetUserClaims(r.Context())
		if !ok || claims.Role != "admin" {
			ErrorResponse(w, http.StatusForbidden, "forbidden: admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetUserClaims(r.Context())
		if !ok || (claims.Role != "analyst" && claims.Role != "admin") {
			ErrorResponse(w, http.StatusForbidden, "forbidden: analyst or admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func ErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "error",
		"message": message,
	})
}