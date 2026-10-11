package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ahobsonsayers/browserfull/internal/proxy"

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

func (s Server) CloseSession(
	_ context.Context, r CloseSessionRequestObject,
) (CloseSessionResponseObject, error) {
	err := s.agentBrowser.Close(r.SessionName)
	if err != nil {
		return nil, err
	}

	return CloseSession204Response{}, nil
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

func (s ServerOverrides) handleConnect(
	w http.ResponseWriter, r *http.Request, sessionName string,
) error {
	info, err := s.server.agentBrowser.Launch(sessionName)
	if err != nil {
		return fmt.Errorf("error launching browser session cdp: %w", err)
	}

	err = proxy.CDP(w, r, info.CDPURL, s.server.allowedOrigins)
	if err != nil {
		return fmt.Errorf("error proxying cdp: %w", err)
	}

	return nil
}
