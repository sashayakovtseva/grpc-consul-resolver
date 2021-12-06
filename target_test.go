package consul

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_newTarget(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name        string
		in          string
		expect      *target
		expectError bool
	}{
		{
			name: "simple",
			in:   "consul://127.0.0.127:8555/my-service",
			expect: &target{
				Addr:       "127.0.0.127:8555",
				Service:    "my-service",
				Near:       "_agent",
				MaxBackoff: time.Second,
			},
		},
		{
			name: "all args",
			in:   "consul://user:password@127.0.0.127:8555/my-service?wait=14s&near=host&insecure=true&limit=1&tag=production&token=test_token&max-backoff=2s&dc=xx&allow-stale=true&require-consistent=true",
			expect: &target{
				Addr:              "127.0.0.127:8555",
				User:              "user",
				Password:          "password",
				Token:             "test_token",
				Wait:              14 * time.Second,
				TLSInsecure:       true,
				AllowStale:        true,
				RequireConsistent: true,
				Dc:                "xx",
				Service:           "my-service",
				Near:              "host",
				MaxBackoff:        2 * time.Second,
				Limit:             1,
				Tag:               "production",
				tags:              []string{"production"},
			},
		},
		{
			name: "multiple tags",
			in:   "consul://user:password@127.0.0.127:8555/my-service?limit=1&tag=production,green&dc=xx",
			expect: &target{
				Addr:       "127.0.0.127:8555",
				User:       "user",
				Password:   "password",
				Dc:         "xx",
				Service:    "my-service",
				Near:       "_agent",
				MaxBackoff: time.Second,
				Limit:      1,
				Tag:        "production,green",
				tags:       []string{"production", "green"},
			},
		},
		{
			name:        "bad scheme",
			in:          "127.0.0.127:8555/my-service",
			expectError: true,
		},
		{
			name:        "no service",
			in:          "consul://127.0.0.127:8555",
			expectError: true,
		},
		{
			name:        "bad arg",
			in:          "consul://127.0.0.127:8555/s?insecure=BADDD",
			expectError: true,
		},
	}

	for i := range tt {
		tc := tt[i]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual, err := newTarget(tc.in)
			require.Equal(t, tc.expectError, err != nil)
			require.Equal(t, tc.expect, actual)
		})
	}
}
