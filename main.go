package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ahobsonsayers/browserfull/api"
	"github.com/ahobsonsayers/browserfull/api/middleware"
	"github.com/ahobsonsayers/browserfull/internal/agentbrowser"
	"github.com/ahobsonsayers/browserfull/internal/config"
	"github.com/ahobsonsayers/browserfull/internal/proxy"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatal(err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Load openapi spec
	openapiSpec, err := api.GetSpec()
	if err != nil {
		log.Fatalf("failed to load openapi spec: %v", err)
	}

	// Create router
	router := chi.NewRouter()
	router.Use(middleware.Logger("browserfull")) // Contains recoverer

	// Create handler
	ab, err := agentbrowser.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Start the dashboard
	err = ab.StartDashboard()
	if err != nil {
		log.Fatalf("failed to start dashboard: %v", err)
	}

	// Register api routes on group so take priority if matched
	router.Group(func(r chi.Router) {
		r.Use(middleware.OpenAPIValidation("/", openapiSpec))
		server := api.NewServer(ab, cfg)
		api.HandlerFromMux(server, r)
	})

	// Serve dashboard only on its known paths
	dashboard := proxy.Dashboard()
	router.Handle("/", dashboard)
	router.Handle("/_next/*", dashboard)
	router.Handle("/api/*", dashboard) // This is the dashboard api
	router.Handle("/favicon.ico", dashboard)

	// Start listening
	address := fmt.Sprintf("0.0.0.0:%d", cfg.Port)
	log.Printf("Server listening on %s\n", address)
	err = http.ListenAndServe(address, router)
	if err != nil {
		log.Fatal(err)
	}
}
