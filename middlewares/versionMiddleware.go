package middlewares

import (
	"net/http"

	"github.com/luponetn/hng-stage-1/utils"
)

// VersionMiddleware ensures the X-API-Version header is present and correct.
func VersionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		version := r.Header.Get("X-API-Version")
		if version != "1" {
			utils.ErrorResponse(w, http.StatusBadRequest, "API version header required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
