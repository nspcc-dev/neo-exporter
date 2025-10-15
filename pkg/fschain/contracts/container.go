package contracts

import (
	"fmt"

	"github.com/nspcc-dev/neo-exporter/pkg/monitor"
	"github.com/nspcc-dev/neo-exporter/pkg/pool"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm/stackitem"
	"github.com/nspcc-dev/neofs-contract/rpc/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
)

type Container struct {
	pool           *pool.Pool
	contractHash   util.Uint160
	contractReader *container.ContractReader
}

// NewContainer creates Container to interact with 'container' contract in morph chain.
func NewContainer(p *pool.Pool, contractHash util.Uint160) (*Container, error) {
	return &Container{
		pool:           p,
		contractHash:   contractHash,
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

// NodeReportSummaries returns summary info about containers.
func (c *Container) NodeReportSummaries() ([]monitor.ContainerInfo, error) {
	inv, err := c.pool.GetIteratorInvoker()
	if err != nil {
		return nil, fmt.Errorf("failed to get invoker: %w", err)
	}

	contractReader := container.NewReader(inv, c.contractHash)

	sid, iter, err := contractReader.IterateAllReportSummaries()
	if err != nil {
		return nil, fmt.Errorf("can't fetch netmap summaries: %w", err)
	}

	var summaries []monitor.ContainerInfo

	for {
		items, err := inv.TraverseIterator(sid, &iter, defaultIteratorPage)
		if err != nil {
			return nil, fmt.Errorf("can't iterate report summaries: %w", err)
		}

		if len(items) == 0 {
			break
		}

		for _, e := range items {
			kv := e.Value().([]stackitem.Item)
			cID, err := kv[0].TryBytes()
			if err != nil {
				return nil, err
			}

			cnrID, err := cid.DecodeBytes(cID)
			if err != nil {
				return nil, err
			}

			v := kv[1].Value().([]stackitem.Item)
			sz, err := v[0].TryInteger()
			if err != nil {
				return nil, err
			}
			objs, err := v[1].TryInteger()
			if err != nil {
				return nil, err
			}

			summaries = append(summaries, monitor.ContainerInfo{
				ID:              cnrID,
				Size:            sz.Uint64(),
				NumberOfObjects: objs.Uint64(),
			})
		}
	}

	return summaries, nil
}
