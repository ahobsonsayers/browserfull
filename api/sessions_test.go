package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ahobsonsayers/browserfull/internal/agentbrowser"
	"github.com/ahobsonsayers/browserfull/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func TestListSessions(t *testing.T) {
	cfg, err := config.Load()
	require.NoError(t, err)

	ab, err := agentbrowser.New(cfg)
	require.NoError(t, err)

	// Launch a real session so the list has something to find
	sessionName := fmt.Sprintf("testsessions%d", time.Now().UnixNano())
	_, err = ab.Launch(sessionName)
	require.NoError(t, err)
	t.Cleanup(func() {
		closeErr := ab.Close(sessionName)
		require.NoError(t, closeErr)
	})

	server := NewServer(ab, cfg)
	router := chi.NewRouter()
	HandlerFromMux(server, router)

	req := httptest.NewRequest(http.MethodGet, "/sessions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var sessions []string
	err = json.Unmarshal(w.Body.Bytes(), &sessions)
	require.NoError(t, err)
	require.Contains(t, sessions, sessionName)
}
