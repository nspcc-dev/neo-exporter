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
		var (
			nodeInfoV2 v2netmap.NodeInfo
			nodeInfo   netmap.NodeInfo
		)

		if err := nodeInfoV2.Unmarshal(d.BLOB); err != nil {
			return nil, fmt.Errorf("can't unmarshal peer info: %w", err)
		}

		switch d.State.Int64() {
		case int64(nodestate.Online):
			nodeInfoV2.SetState(v2netmap.Online)
			nodeInfo.SetOnline()
		case int64(nodestate.Offline):
			nodeInfoV2.SetState(v2netmap.Offline)
			nodeInfo.SetOffline()
		case int64(nodestate.Maintenance):
			nodeInfoV2.SetState(v2netmap.Maintenance)
			nodeInfo.SetMaintenance()
		default:
			nodeInfoV2.SetState(v2netmap.UnspecifiedState)
		}

		var (
			attrs [][2]string
		)

		for _, attribute := range nodeInfoV2.GetAttributes() {
			attrs = append(attrs, [2]string{attribute.GetKey(), attribute.GetValue()})
		}

		nodeInfo.SetAttributes(attrs)

		nodes = append(nodes, &nodeInfo)
	}

	return nodes, nil
}
