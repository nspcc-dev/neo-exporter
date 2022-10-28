package contracts

import (
	"crypto/elliptic"
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/vm/stackitem"
	v2netmap "github.com/nspcc-dev/neofs-api-go/v2/netmap"
	netmapContract "github.com/nspcc-dev/neofs-contract/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
)

func getInt64(st []stackitem.Item) (int64, error) {
	index := len(st) - 1 // top stack element is last in the array
	bi, err := st[index].TryInteger()
	if err != nil {
		return 0, err
	}
	return bi.Int64(), nil
}

func parseNodeInfo(st stackitem.Item) (*netmap.NodeInfo, error) {
	values, ok := st.Value().([]stackitem.Item)
	if !ok {
		return nil, fmt.Errorf("invalid netmap node")
	}

	if len(values) != 2 {
		return nil, fmt.Errorf("invalid netmap node")
	}

	rawNode, err := values[0].TryBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to get node field: %w", err)
	}

	var nodeInfoV2 v2netmap.NodeInfo
	if err = nodeInfoV2.Unmarshal(rawNode); err != nil {
		return nil, fmt.Errorf("can't unmarshal peer info: %w", err)
	}

	state, err := getInt64(values[1:2])
	if err != nil {
		return nil, fmt.Errorf("failed to get state field: %w", err)
	}

	switch state {
	case int64(netmapContract.NodeStateOnline):
		nodeInfoV2.SetState(v2netmap.Online)
	case int64(netmapContract.NodeStateOffline):
		nodeInfoV2.SetState(v2netmap.Offline)
	case int64(netmapContract.NodeStateMaintenance):
		nodeInfoV2.SetState(v2netmap.Maintenance)
	default:
		nodeInfoV2.SetState(v2netmap.UnspecifiedState)
	}

	var nodeInfo netmap.NodeInfo
	if err = nodeInfo.ReadFromV2(nodeInfoV2); err != nil {
		return nil, fmt.Errorf("failed to read node info from v2: %w", err)
	}

	return &nodeInfo, nil
}

func parseIRNode(st stackitem.Item) (*keys.PublicKey, error) {
	values, ok := st.Value().([]stackitem.Item)
	if !ok {
		return nil, fmt.Errorf("invalid ir node")
	}

	if len(values) < 1 {
		return nil, fmt.Errorf("invalid ir node")
	}

	rawKey, err := values[0].TryBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to get node field: %w", err)
	}

	return keys.NewPublicKeyFromBytes(rawKey, elliptic.P256())
}
