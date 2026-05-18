package pool

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/rpcclient"
	"github.com/stretchr/testify/require"
)

func TestIsCurrentHealthyReturnsFalseWhenClientIsNil(t *testing.T) {
	p := &Pool{
		recheckInterval: time.Second,
		clients:         make([]*rpcclient.Client, 1),
	}

	p.lastHealthyTimestamp.Store(time.Now().UTC().UnixNano())

	require.False(t, p.isCurrentHealthy())
}

func TestEstablishNewConnectionAssignsClientToCorrectIndex(t *testing.T) {
	oldNeoGoClient := newNeoGoClient
	t.Cleanup(func() {
		newNeoGoClient = oldNeoGoClient
	})

	goodClient := &rpcclient.Client{}
	newNeoGoClient = func(_ context.Context, endpoint string, _ rpcclient.Options) (*rpcclient.Client, error) {
		if endpoint == "bad" {
			return nil, errors.New("bad endpoint")
		}

		return goodClient, nil
	}

	p := &Pool{
		ctx:       context.Background(),
		endpoints: []string{"bad", "good"},
		opts:      rpcclient.Options{},
		clients:   make([]*rpcclient.Client, 2),
	}

	err := p.establishNewConnection()
	require.NoError(t, err)
	require.Equal(t, 1, p.current)
	require.Equal(t, 0, p.next)
	require.Nil(t, p.clients[0])
	require.Same(t, goodClient, p.clients[1])
}
