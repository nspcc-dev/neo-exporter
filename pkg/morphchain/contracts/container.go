package contracts

import (
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-contract/rpc/container"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/pool"
)

type Container struct {
	contractReader *container.ContractReader
}

// NewContainer creates Container to interact with 'container' contract in morph chain.
func NewContainer(p *pool.Pool, contractHash util.Uint160) (*Container, error) {
	return &Container{
		contractReader: container.NewReader(p, contractHash),
	}, nil
}

func (c *Container) Total() (int64, error) {
	amount, err := c.contractReader.Count()
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return amount.Int64(), nil
}
