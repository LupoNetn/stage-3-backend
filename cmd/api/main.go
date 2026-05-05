package main

import (
	"log"
	"net/http"

	"github.com/luponetn/hng-stage-1/internals/cache"
	"github.com/luponetn/hng-stage-1/internals/config"
	"github.com/luponetn/hng-stage-1/internals/db"
	"github.com/luponetn/hng-stage-1/internals/handlers"
	"github.com/luponetn/hng-stage-1/middlewares"
	"github.com/luponetn/hng-stage-1/utils"
)

func main() {
	// 1. Initialize Configuration
	cfg := config.LoadConfig()

	// 2. Setup Router
	router := http.NewServeMux()

	// 3. Connect to Database
	pool, err := db.ConnectDB(cfg.DBURL)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer pool.Close()

	// 3.5 Connect to Redis
	if cfg.RedisURL != "" {
		if err := cache.InitRedis(cfg.RedisURL); err != nil {
			log.Fatalf("could not connect to redis: %v", err)
		}
	}

	// 4. Initialize Handlers
	queries := db.New(pool)
	h := handlers.NewHandler(queries, pool)		

	//health routes
	router.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		utils.JSONResponse(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	// ==========================================
	// AUTHENTICATION ROUTES
	// ==========================================
	router.HandleFunc("POST /auth/github/cli", h.HandleGithubAuthCallback)
	router.HandleFunc("GET /auth/github", h.HandleGithubAuth)
	router.HandleFunc("GET /auth/github/url", h.HandleGithubAuthURL)
	router.HandleFunc("POST /auth/github/callback", h.HandleGithubAuthCallback)
	router.HandleFunc("GET /auth/github/callback", h.HandleGithubAuthCallback)
	router.Handle("GET /auth/me", 
		middlewares.VersionMiddleware(
			middlewares.AuthMiddleware(http.HandlerFunc(h.HandleMe)),
		),
	)
	// Alias for tester
	router.Handle("GET /api/users/me", 
		middlewares.VersionMiddleware(
			middlewares.AuthMiddleware(http.HandlerFunc(h.HandleMe)),
		),
	)
	router.HandleFunc("POST /auth/refresh", h.HandleRefresh)
	router.HandleFunc("/auth/logout", h.HandleLogout)

	
	
	// Admin Only Routes
	router.Handle("POST /api/profiles", 
		middlewares.VersionMiddleware(
			middlewares.AuthMiddleware(
				middlewares.AuthorizeAdmin(http.HandlerFunc(h.CreateProfile)),
			),
		),
	)

	router.Handle("POST /api/profiles/upload", 
		middlewares.VersionMiddleware(
			middlewares.AuthMiddleware(
				middlewares.AuthorizeAdmin(http.HandlerFunc(h.ImportProfilesCSV)),
			),
		),
	)
	
	router.Handle("DELETE /api/profiles/{id}", 
		middlewares.VersionMiddleware(
			middlewares.AuthMiddleware(
				middlewares.AuthorizeAdmin(http.HandlerFunc(h.DeleteProfileByID)),
			),
		),
	)

	// protected routes
	router.Handle("GET /api/profiles/{id}", 
		middlewares.VersionMiddleware(
			middlewares.AuthMiddleware(
				middlewares.Authorize(http.HandlerFunc(h.GetProfileByID)),
			),
		),
	)
	
	router.Handle("GET /api/profiles", 
		middlewares.VersionMiddleware(
			middlewares.AuthMiddleware(
				middlewares.Authorize(http.HandlerFunc(h.GetProfiles)),
			),
		),
	)
	
	router.Handle("GET /api/profiles/search", 
		middlewares.VersionMiddleware(
			middlewares.AuthMiddleware(
				middlewares.Authorize(http.HandlerFunc(h.SearchProfiles)),
			),
		),
	)

	router.Handle("GET /api/profiles/export", 
		middlewares.VersionMiddleware(
			middlewares.AuthMiddleware(
				middlewares.Authorize(http.HandlerFunc(h.ExportProfiles)),
			),
		),
	)

	
	if err := StartServer(router, cfg.Port); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
