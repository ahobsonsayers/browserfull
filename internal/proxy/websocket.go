package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"

	"github.com/gorilla/websocket"
	"github.com/koding/websocketproxy"
)

func NewUpgrader(allowedOrigins []string) *websocket.Upgrader {
	return &websocket.Upgrader{
		CheckOrigin: newOriginChecker(allowedOrigins),
	}
}

func CDP(
	response http.ResponseWriter,
	request *http.Request,
	cdpURL string,
	allowedOrigins []string,
) error {
	// Validate cdp url
	parsedCDPUrl, err := url.Parse(cdpURL)
	if err != nil {
		return fmt.Errorf("invalid cdp url: %w", err)
	}

	proxy := websocketproxy.NewProxy(parsedCDPUrl)
	proxy.Upgrader = NewUpgrader(allowedOrigins)
	// Strip Origin from the outbound headers so the chrome cdp 
	// endpoint doesn't reject the connection. Out upgrader origin 
	// check will still run against the inbound request.
	proxy.Director = func(_ *http.Request, out http.Header) {
		out.Del("Origin")
	}
	proxy.ServeHTTP(response, request)
	return nil
}

func newOriginChecker(allowedOrigins []string) func(r *http.Request) bool {
	// Check origin function that allows all origins
	if slices.Contains(allowedOrigins, "0.0.0.0") {
		return func(_ *http.Request) bool { return true }
	}

	// Check origin function which is a modified version of
	// default origin check function in gorilla/websocket
	return func(request *http.Request) bool {
		origin := request.Header.Get("Origin")
		if origin == "" {
			return true
		}
		originUrl, err := url.Parse(origin)
		if err != nil {
			return false
		}

		// Allow if origin hostname equals the request hostname,
		// or it is in the allows origin hostname
		return originUrl.Hostname() == request.URL.Hostname() ||
			slices.Contains(allowedOrigins, originUrl.Hostname())
	}
}
