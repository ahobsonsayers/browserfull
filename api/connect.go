package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ahobsonsayers/browserful/internal/proxy"

	haikunator "github.com/atrox/haikunatorgo/v2"
)

var sessionNameGenerator = haikunator.New()

// ConnectDefaultSession is overridden by the ServerOverrides ConnectDefaultSession below
func (Server) ConnectDefaultSession(
	context.Context, ConnectDefaultSessionRequestObject,
) (ConnectDefaultSessionResponseObject, error) {
	return nil, nil
}

// ConnectNamedSession is overridden by the ServerOverrides ConnectNamedSession below
func (Server) ConnectNamedSession(
	context.Context, ConnectNamedSessionRequestObject,
) (ConnectNamedSessionResponseObject, error) {
	return nil, nil
}

// CloseSession is overridden by the ServerOverrides CloseSession below
func (Server) CloseSession(
	context.Context, CloseSessionRequestObject,
) (CloseSessionResponseObject, error) {
	return nil, nil
}

func (s ServerOverrides) ConnectDefaultSession(w http.ResponseWriter, r *http.Request) {
	sessionName := sessionNameGenerator.Haikunate()
	err := s.handleConnect(w, r, sessionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s ServerOverrides) ConnectNamedSession(w http.ResponseWriter, r *http.Request, sessionName string) {
	err := s.handleConnect(w, r, sessionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s ServerOverrides) CloseSession(w http.ResponseWriter, _ *http.Request, sessionName string) {
	err := s.agentBrowser.Close(sessionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s ServerOverrides) handleConnect(
	w http.ResponseWriter, r *http.Request, sessionName string,
) error {
	info, err := s.agentBrowser.Launch(sessionName)
	if err != nil {
		return fmt.Errorf("error launching browser session cdp: %w", err)
	}

	err = proxy.CDP(w, r, info.CDPURL, s.allowedOrigins)
	if err != nil {
		return fmt.Errorf("error proxying cdp: %w", err)
	}

	return nil
}
