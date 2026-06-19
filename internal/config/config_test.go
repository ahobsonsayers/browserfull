package config

import (
	"os"
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
			cfg:     Config{Port: 8080, SessionsDir: "data/sessions"},
			wantErr: false,
		},
		{
			name:    "empty sessions dir",
			cfg:     Config{Port: 8080, SessionsDir: ""},
			wantErr: true,
		},
		{
			name:    "zero port",
			cfg:     Config{Port: 0, SessionsDir: "data/sessions"},
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
	os.Unsetenv("BROWSERFUL_PORT")
	os.Unsetenv("BROWSERFUL_SESSIONS_DIR")

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, uint16(8080), cfg.Port)
	assert.Equal(t, "data/sessions", cfg.SessionsDir)
}

func TestLoad_Override(t *testing.T) {
	t.Setenv("BROWSERFUL_PORT", "9090")
	t.Setenv("BROWSERFUL_SESSIONS_DIR", "/tmp/sessions")

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, uint16(9090), cfg.Port)
	assert.Equal(t, "/tmp/sessions", cfg.SessionsDir)
}

func TestLoad_InvalidPort(t *testing.T) {
	t.Setenv("BROWSERFUL_PORT", "not-a-number")

	cfg, err := Load()
	require.Error(t, err)
	assert.Nil(t, cfg)
}
