package contracts

import (
	"fmt"

	v2netmap "github.com/nspcc-dev/neofs-api-go/v2/netmap"
	"github.com/nspcc-dev/neofs-contract/contracts/netmap/nodestate"
	rcpnetmap "github.com/nspcc-dev/neofs-contract/rpc/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
)

func parseContractNodes(data []*rcpnetmap.NetmapNode) ([]*netmap.NodeInfo, error) {
	nodes := make([]*netmap.NodeInfo, 0, len(data))
	for _, d := range data {
		var nodeInfoV2 v2netmap.NodeInfo
		if err := nodeInfoV2.Unmarshal(d.BLOB); err != nil {
			return nil, fmt.Errorf("can't unmarshal peer info: %w", err)
		}

		switch d.State.Int64() {
		case int64(nodestate.Online):
			nodeInfoV2.SetState(v2netmap.Online)
		case int64(nodestate.Offline):
			nodeInfoV2.SetState(v2netmap.Offline)
		case int64(nodestate.Maintenance):
			nodeInfoV2.SetState(v2netmap.Maintenance)
		default:
			nodeInfoV2.SetState(v2netmap.UnspecifiedState)
		}

		var nodeInfo netmap.NodeInfo
		if err := nodeInfo.ReadFromV2(nodeInfoV2); err != nil {
			return nil, fmt.Errorf("failed to read node info from v2: %w", err)
		}

		nodes = append(nodes, &nodeInfo)
	}

	return nodes, nil
}
