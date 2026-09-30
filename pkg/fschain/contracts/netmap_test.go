package contracts

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEndpointHost(t *testing.T) {
	for _, tc := range []struct {
		address  string
		expected string
		err      bool
	}{
		{
			address:  "[2004:eb1::1]:8080",
			expected: "2004:eb1::1",
		},
		{
			address:  "grpcs://example.com:7070",
			expected: "example.com",
		},
		{
			address:  "172.16.14.1:8080",
			expected: "172.16.14.1",
		},
		{
			address:  "s01.neofs.devenv:8080",
			expected: "s01.neofs.devenv",
		},
		{
			address:  "localhost:8080",
			expected: "localhost",
		},
		{
			address:  "grpcs://s04.neofs.devenv:8082",
			expected: "s04.neofs.devenv",
		},
		{
			address:  "grpcs://[2004:eb1::1]:8080",
			expected: "2004:eb1::1",
		},
		{
			address:  "grpc://172.16.14.1:8080",
			expected: "172.16.14.1",
		},
		{
			address: "http://172.16.14.1:8080",
			err:     true,
		},
		{
			address: "/ip4/172.16.14.1/tcp/8080",
			err:     true,
		},
		{
			address: "/dns4/s04.neofs.devenv/tcp/8082/tls",
			err:     true,
		},
		{
			address: "",
			err:     true,
		},
	} {
		t.Run(tc.address, func(t *testing.T) {
			host, err := endpointHost(tc.address)
			if tc.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, host)
			}
		})
	}
}
