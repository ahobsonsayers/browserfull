package agentbrowser

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ahobsonsayers/browserful/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentBrowserLaunch(t *testing.T) {
	dataDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dataDir) })

	cfg := &config.Config{
		Port:    8080,
		DataDir: dataDir,
	}

	browser, err := New(cfg)
	require.NoError(t, err)

	sessionName := fmt.Sprintf("%d", time.Now().UnixNano())

	info, err := browser.Launch(sessionName)
	require.NoError(t, err)
	require.NotNil(t, info)

	t.Cleanup(func() { browser.Close(sessionName) })

	defer assert.Equal(t, sessionName, info.SessionName)
	assert.NotEmpty(t, info.CDPURL)
	assert.Positive(t, info.PID)
	assert.NotEmpty(t, info.Engine)
	assert.NotEmpty(t, info.Version)
}
