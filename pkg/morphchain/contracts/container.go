package contracts

import (
	"github.com/nspcc-dev/neo-go/pkg/smartcontract"
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
	res, err := c.pool.InvokeFunction(c.contractHash, count, []smartcontract.Parameter{}, nil)
	if err != nil {
		return 0, err
	}

	if err = getInvocationError(res); err != nil {
		return 0, err
	}

	return getInt64(res.Stack)
}
