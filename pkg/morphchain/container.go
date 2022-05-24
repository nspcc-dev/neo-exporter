package morphchain

import (
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/pool"
	"go.uber.org/zap"
)

type (
	// ContainerFetcher is a client to container contract.
	ContainerFetcher struct {
		cli    *pool.ContainerPool
		logger *zap.Logger
	}

	// ContainerFetcherArgs contains parameters to create ContainerFetcher.
	ContainerFetcherArgs struct {
		Cli               *pool.Pool
		Key               *keys.PrivateKey
		ContainerContract util.Uint160
		Logger            *zap.Logger
	}
)

// NewContainerFetcher returns client to communicate
// with container contract in side chain.
func NewContainerFetcher(p ContainerFetcherArgs) (*ContainerFetcher, error) {
	cnrPool, err := pool.NewContainerPool(p.Cli, p.Key, p.ContainerContract)
	if err != nil {
		return nil, fmt.Errorf("can't create container client wrapper: %w", err)
	}

	return &ContainerFetcher{
		cli:    cnrPool,
		logger: p.Logger,
	}, nil
}

// Total returns number of containers available in the network.
func (f *ContainerFetcher) Total() (int, error) {
	cids, err := f.cli.Containers()
	if err != nil {
		return 0, fmt.Errorf("can't fetch list of containers: %w", err)
	}

	return len(cids), nil
}
