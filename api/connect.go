package api

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/ahobsonsayers/browserfull/internal/agentbrowser"
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

func (s ServerOverrides) ConnectDefaultSession(
	w http.ResponseWriter, r *http.Request, params ConnectDefaultSessionParams,
) {
	sessionName := sessionNameGenerator.Haikunate()
	s.connect(w, r, sessionName, params.Headed, params.Extensions)
}

func (s ServerOverrides) ConnectNamedSession(
	w http.ResponseWriter, r *http.Request, sessionName string, params ConnectNamedSessionParams,
) {
	s.connect(w, r, sessionName, params.Headed, params.Extensions)
}

func (s ServerOverrides) connect(
	w http.ResponseWriter,
	r *http.Request,
	sessionName string,
	headed *bool,
	extensions *Extensions,
) {
	extensionPaths, err := s.server.launchExtensions(extensions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.handleConnect(w, r, sessionName, headed, extensionPaths)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s ServerOverrides) handleConnect(
	w http.ResponseWriter,
	r *http.Request,
	sessionName string,
	headed *bool,
	extensions []string,
) error {
	opts := agentbrowser.LaunchOptions{
		Headed:     s.server.headed(headed),
		Extensions: extensions,
	}

	sessionInfo, err := s.server.agentBrowser.Launch(sessionName, opts)
	if err != nil {
		return fmt.Errorf("error launching browser session cdp: %w", err)
	}

	err = proxy.CDP(w, r, sessionInfo.CDPURL, s.server.allowedOrigins)
	if err != nil {
		return fmt.Errorf("error proxying cdp: %w", err)
	}

	return nil
}

func (s Server) headed(requested *bool) bool {
	if requested != nil {
		return *requested
	}

	return s.defaultHeaded
}

func (s Server) launchExtensions(requested *Extensions) ([]string, error) {
	names := s.cfg.Extensions
	if requested != nil {
		names = slices.Concat(s.cfg.Extensions, *requested)
	}

	seen := make(map[string]struct{}, len(names))
	paths := make([]string, 0, len(names))

	for _, name := range names {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}

		path, err := s.cfg.ExtensionPath(name)
		if err != nil {
			return nil, err
		}

		paths = append(paths, path)
	}

	return paths, nil
}
