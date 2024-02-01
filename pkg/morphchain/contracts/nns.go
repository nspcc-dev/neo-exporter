package contracts

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neo-exporter/pkg/pool"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-contract/rpc/nns"
)

type (
	NNS struct {
		contractReader *nns.ContractReader
	}

	NNSNoOp struct {
	}
)

// NewNNS creates NNS to interact with 'nns' contract in morph chain.
func NewNNS(p *pool.Pool, contractHash util.Uint160) (*NNS, error) {
	return &NNS{
		contractReader: nns.NewReader(p, contractHash),
	}, nil
}

func (c *NNS) ResolveFSContract(name string) (util.Uint160, error) {
	hash, err := c.contractReader.ResolveFSContract(name)
	if err != nil {
		return util.Uint160{}, fmt.Errorf("ResolveFSContract: %w", err)
	}

	return hash, nil
}

func (c *NNSNoOp) ResolveFSContract(_ string) (util.Uint160, error) {
	return util.Uint160{}, errors.New("no op")
}
