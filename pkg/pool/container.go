package pool

import (
	"fmt"
	"sync"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/util"
	cid "github.com/nspcc-dev/neofs-api-go/pkg/container/id"
	morph "github.com/nspcc-dev/neofs-node/pkg/morph/client"
	"github.com/nspcc-dev/neofs-node/pkg/morph/client/container/wrapper"
)

// ContainerPool is a wrapper over container contract.
// Similar to wrapper.Wrapper but it accepts pool instead of single rpc client.
type ContainerPool struct {
	*Pool

	mu      sync.RWMutex
	wrapper *wrapper.Wrapper

	key       *keys.PrivateKey
	container util.Uint160
}

// NewContainerPool creates new pool wrapper using parameters.
func NewContainerPool(p *Pool, key *keys.PrivateKey, containerContract util.Uint160) (*ContainerPool, error) {
	conn, _, err := p.nextConnection()
	if err != nil {
		return nil, err
	}

	wrapperNetmap, err := newContainerWrapper(key, conn, containerContract)
	if err != nil {
		return nil, err
	}

	return &ContainerPool{
		Pool:      p,
		wrapper:   wrapperNetmap,
		container: containerContract,
		key:       key,
	}, nil
}

// Containers returns list of all available container IDs in the network.
func (p *ContainerPool) Containers() ([]*cid.ID, error) {
	wrap, err := p.nextWrapper()
	if err != nil {
		return nil, err
	}

	return wrap.List(nil)
}

func newContainerWrapper(key *keys.PrivateKey, conn *client.Client, containerContract util.Uint160) (*wrapper.Wrapper, error) {
	blockchainClient, err := morph.New(
		key,
		"", // endpoint is ignored due to single client instance
		morph.WithSingleClient(conn),
	)
	if err != nil {
		return nil, fmt.Errorf("can't create blockchain client: %w", err)
	}

	return wrapper.NewFromMorph(
		blockchainClient,
		containerContract,
		0,
	)
}

func (p *ContainerPool) nextWrapper() (*wrapper.Wrapper, error) {
	conn, updated, err := p.nextConnection()
	if err != nil {
		return nil, err
	}

	// That's why the lock isn't inside the 'if' branch below
	// https://github.com/nspcc-dev/neofs-net-monitor/pull/64#discussion_r837242673
	p.mu.Lock()
	defer p.mu.Unlock()

	if updated {
		p.wrapper, err = newContainerWrapper(p.key, conn, p.container)
		if err != nil {
			return nil, err
		}
	}

	return p.wrapper, nil
}
