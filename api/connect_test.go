package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ahobsonsayers/browserfull/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerHeaded(t *testing.T) {
	tests := []struct {
		name         string
		defaultValue bool
		requested    *bool
		want         bool
	}{
		{name: "default off", defaultValue: false, want: false},
		{name: "default on", defaultValue: true, want: true},
		{name: "request off overrides default on", defaultValue: true, requested: boolPtr(false), want: false},
		{name: "request on overrides default off", defaultValue: false, requested: boolPtr(true), want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := Server{defaultHeaded: test.defaultValue}
			assert.Equal(t, test.want, server.headed(test.requested))
		})
	}
}

func boolPtr(value bool) *bool {
	return &value
}

func TestServerLaunchExtensions(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	for _, name := range []string{"ublock", "grammarly"} {
		dir := filepath.Join(tmpDir, "extensions", name)
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "manifest.json"), []byte("{}"), 0o644))
	}

	server := Server{cfg: &config.Config{DataDir: tmpDir, Extensions: []string{"ublock"}}}

	requested := Extensions{"grammarly", "ublock"}

	paths, err := server.launchExtensions(&requested)
	require.NoError(t, err)
	assert.Equal(t, []string{
		filepath.Join(tmpDir, "extensions", "ublock"),
		filepath.Join(tmpDir, "extensions", "grammarly"),
	}, paths)

	defaults, err := server.launchExtensions(nil)
	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(tmpDir, "extensions", "ublock")}, defaults)

	missing := Extensions{"missing"}
	_, err = server.launchExtensions(&missing)
	require.Error(t, err)
}
