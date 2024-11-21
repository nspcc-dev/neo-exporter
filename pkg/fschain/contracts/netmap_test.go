package contracts

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMultiAddrToIPStringWithoutPort(t *testing.T) {
	for _, tc := range []struct {
		address  string
		expected string
		err      bool
	}{
		{
			address:  "/ip4/172.16.14.1/tcp/8080",
			expected: "172.16.14.1",
		},
		{
			address:  "/dns4/neofs.bigcorp.com/tcp/8080",
			expected: "neofs.bigcorp.com",
		},
		{
			address:  "/dns4/s04.neofs.devenv/tcp/8082/tls",
			expected: "s04.neofs.devenv",
		},
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
			address:  "http://172.16.14.1:8080",
			expected: "172.16.14.1",
			err:      true,
		},
	} {
		t.Run("", func(t *testing.T) {
			host, err := multiAddrToIPStringWithoutPort(tc.address)
			if tc.err {
				require.Error(t, err)
			} else {
				require.Equal(t, tc.expected, host)
			}
		})
	}
}
