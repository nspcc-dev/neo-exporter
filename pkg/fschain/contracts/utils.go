package contracts

import (
	"fmt"

	"github.com/nspcc-dev/neofs-contract/contracts/netmap/nodestate"
	rcpnetmap "github.com/nspcc-dev/neofs-contract/rpc/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
)

func parseContractNodes(data []*rcpnetmap.NetmapNode) ([]*netmap.NodeInfo, error) {
	nodes := make([]*netmap.NodeInfo, 0, len(data))
	for _, d := range data {
		var (
			nodeInfo netmap.NodeInfo
		)

		if err := nodeInfo.Unmarshal(d.BLOB); err != nil {
			return nil, fmt.Errorf("can't unmarshal peer info: %w", err)
		}

		switch d.State.Int64() {
		case int64(nodestate.Online):
			nodeInfo.SetOnline()
		case int64(nodestate.Offline):
			nodeInfo.SetOffline()
		case int64(nodestate.Maintenance):
			nodeInfo.SetMaintenance()
		}

		nodes = append(nodes, &nodeInfo)
	}

	return nodes, nil
}
