package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "valid defaults",
			cfg:     Config{Port: 8080, DataDir: "/tmp/browserfull"},
			wantErr: false,
		},
		{
			name:    "valid high port",
			cfg:     Config{Port: 65535, DataDir: "/tmp/browserfull"},
			wantErr: false,
		},
		{
			name:    "zero port",
			cfg:     Config{Port: 0, DataDir: "/tmp/browserfull"},
			wantErr: true,
		},
		{
			name:    "empty data dir",
			cfg:     Config{Port: 8080, DataDir: ""},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.cfg.Validate()
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLoad_Defaults(t *testing.T) {
	os.Unsetenv("BROWSERFULL_PORT")
	os.Unsetenv("BROWSERFULL_DATA_DIR")

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, uint16(8080), cfg.Port)
	assert.NotEmpty(t, cfg.DataDir)
}

func TestLoad_Override(t *testing.T) {
	t.Setenv("BROWSERFULL_PORT", "9090")
	t.Setenv("BROWSERFULL_DATA_DIR", "/tmp/browserfull-data")

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, uint16(9090), cfg.Port)
	assert.Equal(t, "/tmp/browserfull-data", cfg.DataDir)
}

func TestLoad_InvalidPort(t *testing.T) {
	t.Setenv("BROWSERFULL_PORT", "not-a-number")

	cfg, err := Load()
	require.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoad_AllowedOrigins(t *testing.T) {
	t.Run("unset defaults to empty", func(t *testing.T) {
		os.Unsetenv("BROWSERFULL_ALLOWED_ORIGINS")

		cfg, err := Load()
		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Empty(t, cfg.AllowedOrigins)
	})

	t.Run("single value", func(t *testing.T) {
		t.Setenv("BROWSERFULL_ALLOWED_ORIGINS", "*")

		cfg, err := Load()
		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, []string{"*"}, cfg.AllowedOrigins)
	})

	t.Run("comma separated", func(t *testing.T) {
		t.Setenv("BROWSERFULL_ALLOWED_ORIGINS", "127.0.0.1,10.0.0.5")

		cfg, err := Load()
		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, []string{"127.0.0.1", "10.0.0.5"}, cfg.AllowedOrigins)
	})
}

func TestExtensionPath(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	ublockDir := filepath.Join(tmpDir, "extensions", "ublock")
	require.NoError(t, os.MkdirAll(ublockDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(ublockDir, "manifest.json"), []byte("{}"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "extensions", "nomanifest"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "extensions", "file"), []byte("{}"), 0o644))

	cfg := Config{Port: 8080, DataDir: tmpDir}

	path, err := cfg.ExtensionPath("ublock")
	require.NoError(t, err)
	assert.Equal(t, ublockDir, path)

	for _, name := range []string{"", ".", "..", "missing", "nomanifest", "file", "../ublock", "a/b"} {
		_, err = cfg.ExtensionPath(name)
		assert.Error(t, err, name)
	}
}

func TestValidate_Extensions(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	ublockDir := filepath.Join(tmpDir, "extensions", "ublock")
	require.NoError(t, os.MkdirAll(ublockDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(ublockDir, "manifest.json"), []byte("{}"), 0o644))

	valid := Config{Port: 8080, DataDir: tmpDir, Extensions: []string{"ublock"}}
	require.NoError(t, valid.Validate())

	invalid := Config{Port: 8080, DataDir: tmpDir, Extensions: []string{"missing"}}
	require.Error(t, invalid.Validate())
}
