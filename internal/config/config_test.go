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
			cfg:     Config{Port: 8080, DataDir: "/tmp/browserful"},
			wantErr: false,
		},
		{
			name:    "valid high port",
			cfg:     Config{Port: 65535, DataDir: "/tmp/browserful"},
			wantErr: false,
		},
		{
			name:    "zero port",
			cfg:     Config{Port: 0, DataDir: "/tmp/browserful"},
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
	os.Unsetenv("BROWSERFUL_PORT")
	os.Unsetenv("BROWSERFUL_DATA_DIR")

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, uint16(8080), cfg.Port)
	assert.NotEmpty(t, cfg.DataDir)
}

func TestLoad_Override(t *testing.T) {
	t.Setenv("BROWSERFUL_PORT", "9090")
	t.Setenv("BROWSERFUL_DATA_DIR", "/tmp/browserful-data")

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, uint16(9090), cfg.Port)
	assert.Equal(t, "/tmp/browserful-data", cfg.DataDir)
}

func TestLoad_InvalidPort(t *testing.T) {
	t.Setenv("BROWSERFUL_PORT", "not-a-number")

	cfg, err := Load()
	require.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoad_AllowedOrigins(t *testing.T) {
	t.Run("unset defaults to empty", func(t *testing.T) {
		os.Unsetenv("BROWSERFUL_ALLOWED_ORIGINS")

		cfg, err := Load()
		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Empty(t, cfg.AllowedOrigins)
	})

	t.Run("single value", func(t *testing.T) {
		t.Setenv("BROWSERFUL_ALLOWED_ORIGINS", "*")

		cfg, err := Load()
		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, []string{"*"}, cfg.AllowedOrigins)
	})

	t.Run("comma separated", func(t *testing.T) {
		t.Setenv("BROWSERFUL_ALLOWED_ORIGINS", "127.0.0.1,10.0.0.5")

		cfg, err := Load()
		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, []string{"127.0.0.1", "10.0.0.5"}, cfg.AllowedOrigins)
	})
}
