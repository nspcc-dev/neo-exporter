package pool

import (
	"fmt"
	"sync"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-api-go/pkg/netmap"
	morph "github.com/nspcc-dev/neofs-node/pkg/morph/client"
	"github.com/nspcc-dev/neofs-node/pkg/morph/client/netmap/wrapper"
)

// NetmapPool is a wrapper over netmap contract.
// Similar to wrapper.Wrapper but it accepts pool instead of single rpc client.
type NetmapPool struct {
	*Pool

	mu      sync.RWMutex
	wrapper *wrapper.Wrapper

	key    *keys.PrivateKey
	netmap util.Uint160
}

// NewNetmapPool creates new pool wrapper using parameters.
func NewNetmapPool(p *Pool, key *keys.PrivateKey, netmapContract util.Uint160) (*NetmapPool, error) {
	conn, _, err := p.nextConnection()
	if err != nil {
		return nil, err
	}

	wrapperNetmap, err := newNetmapWrapper(key, conn, netmapContract)
	if err != nil {
		return nil, err
	}

	return &NetmapPool{
		Pool:    p,
		wrapper: wrapperNetmap,
		netmap:  netmapContract,
		key:     key,
	}, nil
}

// Epoch receives number of current NeoFS epoch
// through the Netmap contract call.
func (p *NetmapPool) Epoch() (uint64, error) {
	wrap, err := p.nextWrapper()
	if err != nil {
		return 0, err
	}

	return wrap.Epoch()
}

// GetCandidates receives information list about candidates
// for the next epoch network map through the Netmap contract
// call, composes network map from them and returns it.
func (p *NetmapPool) GetCandidates() (*netmap.Netmap, error) {
	wrap, err := p.nextWrapper()
	if err != nil {
		return nil, err
	}

	return wrap.GetCandidates()
}

// GetNetMap receives information list about storage nodes
// through the Netmap contract call, composes network map
// from them and returns it. With diff == 0 returns current
// network map, else return snapshot of previous network map.
func (p *NetmapPool) GetNetMap() (*netmap.Netmap, error) {
	wrap, err := p.nextWrapper()
	if err != nil {
		return nil, err
	}

	return wrap.GetNetMap(0)
}

// GetInnerRingList return current IR list.
func (p *NetmapPool) GetInnerRingList() (keys.PublicKeys, error) {
	wrap, err := p.nextWrapper()
	if err != nil {
		return nil, err
	}

	return wrap.GetInnerRingList()
}

// NeoFSAlphabetList returns keys that stored in NeoFS Alphabet role. Main chain
// stores alphabet node keys of inner ring there, however side chain stores both
// alphabet and non alphabet node keys of inner ring.
func (p *NetmapPool) NeoFSAlphabetList() (keys.PublicKeys, error) {
	wrap, err := p.nextWrapper()
	if err != nil {
		return nil, err
	}

	return wrap.Morph().NeoFSAlphabetList()
}

func newNetmapWrapper(key *keys.PrivateKey, conn *client.Client, netmapContract util.Uint160) (*wrapper.Wrapper, error) {
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
		netmapContract,
		0,
	)
}

func (p *NetmapPool) nextWrapper() (*wrapper.Wrapper, error) {
	conn, updated, err := p.nextConnection()
	if err != nil {
		return nil, err
	}

	// That's why the lock isn't inside the 'if' branch below
	// https://github.com/nspcc-dev/neofs-net-monitor/pull/64#discussion_r837242673
	p.mu.Lock()
	defer p.mu.Unlock()

	if updated {
		p.wrapper, err = newNetmapWrapper(p.key, conn, p.netmap)
		if err != nil {
			return nil, err
		}
	}

	return p.wrapper, nil
}
