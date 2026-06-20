package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUpgrader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		allowedOrigins []string
		originHeader   string
		requestHost    string
		want           bool
	}{
		{
			name:           "empty allowlist: same origin allowed",
			allowedOrigins: nil,
			originHeader:   "http://127.0.0.1:8080",
			requestHost:    "127.0.0.1:8080",
			want:           true,
		},
		{
			name:           "empty allowlist: cross origin rejected",
			allowedOrigins: nil,
			originHeader:   "http://evil.com",
			requestHost:    "127.0.0.1:8080",
			want:           false,
		},
		{
			name:           "empty allowlist: missing origin allowed",
			allowedOrigins: nil,
			originHeader:   "",
			requestHost:    "127.0.0.1:8080",
			want:           true,
		},
		{
			name:           "wildcard: cross origin allowed",
			allowedOrigins: []string{"*"},
			originHeader:   "http://evil.com",
			requestHost:    "127.0.0.1:8080",
			want:           true,
		},
		{
			name:           "wildcard: missing origin allowed",
			allowedOrigins: []string{"*"},
			originHeader:   "",
			requestHost:    "127.0.0.1:8080",
			want:           true,
		},
		{
			name:           "allowlist: matching hostname allowed",
			allowedOrigins: []string{"127.0.0.1", "10.0.0.5"},
			originHeader:   "http://10.0.0.5:9090",
			requestHost:    "127.0.0.1:8080",
			want:           true,
		},
		{
			name:           "allowlist: non-matching hostname rejected",
			allowedOrigins: []string{"127.0.0.1"},
			originHeader:   "http://evil.com",
			requestHost:    "127.0.0.1:8080",
			want:           false,
		},
		{
			name:           "allowlist: same origin allowed even when host not in list",
			allowedOrigins: []string{"10.0.0.5"},
			originHeader:   "http://127.0.0.1:8080",
			requestHost:    "127.0.0.1:8080",
			want:           true,
		},
		{
			name:           "allowlist: missing origin allowed",
			allowedOrigins: []string{"127.0.0.1"},
			originHeader:   "",
			requestHost:    "127.0.0.1:8080",
			want:           true,
		},
		{
			name:           "allowlist: invalid origin url rejected",
			allowedOrigins: []string{"127.0.0.1"},
			originHeader:   "://not-a-url",
			requestHost:    "127.0.0.1:8080",
			want:           false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			upgrader := NewUpgrader(tc.allowedOrigins)
			require.NotNil(t, upgrader)
			require.NotNil(t, upgrader.CheckOrigin)

			req := httptest.NewRequest(http.MethodGet, "http://"+tc.requestHost+"/", http.NoBody)
			if tc.originHeader != "" {
				req.Header.Set("Origin", tc.originHeader)
			}
			req.Host = tc.requestHost

			got := upgrader.CheckOrigin(req)
			assert.Equal(t, tc.want, got)
		})
	}
}
