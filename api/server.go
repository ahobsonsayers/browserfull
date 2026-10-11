package api

import (
	"github.com/ahobsonsayers/browserfull/internal/agentbrowser"
	"github.com/ahobsonsayers/browserfull/internal/config"
)

// Server is the type that all handlers are defined on as methods
// It holds the service state and satisfies the strict server interface
type Server struct {
	agentBrowser   *agentbrowser.Client
	allowedOrigins []string
}

var _ StrictServerInterface = Server{}

// ServerOverrides is a type that handler overrides (with access
// to request and response types) can be defined on.
// It satisfies the service interface by embedding another server interface
type ServerOverrides struct {
	ServerInterface
	server Server
}

var _ ServerInterface = ServerOverrides{}

func NewServer(agentBrowser *agentbrowser.Client, cfg *config.Config) ServerInterface {
	// Create strict service
	server := Server{
		agentBrowser:   agentBrowser,
		allowedOrigins: cfg.AllowedOrigins,
	}

	// Convert to none strict server
	strictServer := NewStrictHandler(server, nil)

	// Apply server overrides
	overriddenServer := ServerOverrides{
		// strictServer wraps server's methods as standard http.Handler calls;
		// server is kept for overrides that need the raw ResponseWriter/Request
		ServerInterface: strictServer,
		server:          server,
	}

	return overriddenServer
}
