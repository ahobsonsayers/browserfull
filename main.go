package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ahobsonsayers/browserful/api"
	"github.com/ahobsonsayers/browserful/api/middleware"
	"github.com/ahobsonsayers/browserful/internal/agentbrowser"
	"github.com/ahobsonsayers/browserful/internal/config"
	"github.com/go-chi/chi/v5"
)

func main() {
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
	router.Use(middleware.Logger("browserful")) // Contains recoverer
	router.Use(middleware.OpenAPIValidation("/", openapiSpec))

	// Create handler
	ab, err := agentbrowser.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	server := api.NewServer(ab, cfg)
	handler := api.HandlerFromMux(server, router)

	// Start listening
	address := fmt.Sprintf("0.0.0.0:%d", cfg.Port)
	log.Printf("Server listening on %s\n", address)
	err = http.ListenAndServe(address, handler)
	if err != nil {
		log.Fatal(err)
	}
}
