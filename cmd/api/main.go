package main

import (
	"log"

	"github.com/luponetn/hng-stage-1/internals/config"
	"github.com/luponetn/hng-stage-1/internals/db"
	"github.com/luponetn/hng-stage-1/internals/handlers"
)

func main() {
	cfg := config.LoadConfig()

	router := CreateRouter()

	//connect with db
	pool, err := db.ConnectDB(cfg.DBURL)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer pool.Close()

	queries := db.New(pool)

	h := handlers.NewHandler(queries)

	//auth endpoints
	router.HandleFunc("POST /auth/github/cli", h.HandleGithubCLIAuth)
	router.HandleFunc("GET /auth/github", h.HandleGithubAuth)
	router.HandleFunc("POST /auth/github/callback", h.HandleGithubAuthCallback)
	router.HandleFunc("GET /auth/me", h.HandleMe)
	router.HandleFunc("POST /auth/refresh", h.HandleRefresh)
	router.HandleFunc("POST /auth/logout", h.HandleLogout)

	router.HandleFunc("POST /api/profiles", h.CreateProfile)
	router.HandleFunc("GET /api/profiles/{id}", h.GetProfileByID)
	router.HandleFunc("GET /api/profiles", h.GetProfiles)
	router.HandleFunc("GET /api/profiles/search", h.SearchProfiles)
	router.HandleFunc("DELETE /api/profiles/{id}", h.DeleteProfileByID)

	if err := StartServer(router, cfg.Port); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
