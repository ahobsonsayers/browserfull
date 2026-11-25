package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
