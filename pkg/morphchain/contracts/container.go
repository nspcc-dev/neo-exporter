package contracts

import (
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/unwrap"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/pool"
)

type Container struct {
	pool         *pool.Pool
	contractHash util.Uint160
}

const (
	count = "count"
)

// NewContainer creates Container to interact with 'container' contract in morph chain.
func NewContainer(p *pool.Pool, contractHash util.Uint160) (*Container, error) {
	return &Container{
		pool:         p,
		contractHash: contractHash,
	}, nil
}

func (c *Container) Total() (int64, error) {
	return unwrap.Int64(c.pool.Call(c.contractHash, count))
}
