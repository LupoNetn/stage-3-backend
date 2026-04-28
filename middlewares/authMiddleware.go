package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/luponetn/hng-stage-1/utils"
)

type contextKey string

const UserContextKey contextKey = "user"

// GetUserClaims retrieves the JWT claims from the request context.
func GetUserClaims(ctx context.Context) (*utils.Claims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*utils.Claims)
	return claims, ok
}

// AuthMiddleware validates the JWT token from cookies or the Authorization header.
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
			utils.ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		claims, err := utils.VerifyToken(tokenString)
		if err != nil {
			utils.ErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthorizeAdmin ensures the authenticated user has an 'admin' role.
func AuthorizeAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetUserClaims(r.Context())
		if !ok || claims.Role != "admin" {
			utils.ErrorResponse(w, http.StatusForbidden, "forbidden: admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Authorize ensures the authenticated user has either 'analyst' or 'admin' role.
func Authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetUserClaims(r.Context())
		if !ok || (claims.Role != "analyst" && claims.Role != "admin") {
			utils.ErrorResponse(w, http.StatusForbidden, "forbidden: analyst or admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
