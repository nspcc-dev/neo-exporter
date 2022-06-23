package contracts

import (
	"crypto/elliptic"
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/rpc/response/result"
	"github.com/nspcc-dev/neo-go/pkg/vm"
	"github.com/nspcc-dev/neo-go/pkg/vm/stackitem"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
)

func getInvocationError(result *result.Invoke) error {
	if result.State != vm.HaltState.String() {
		return fmt.Errorf("invocation failed: %s", result.FaultException)
	}
	if len(result.Stack) == 0 {
		return fmt.Errorf("result stack is empty")
	}
	return nil
}

func getInt64(st []stackitem.Item) (int64, error) {
	index := len(st) - 1 // top stack element is last in the array
	bi, err := st[index].TryInteger()
	if err != nil {
		return 0, err
	}
	return bi.Int64(), nil
}

func parseNode(st stackitem.Item) (*netmap.NodeInfo, error) {
	values, ok := st.Value().([]stackitem.Item)
	if !ok {
		return nil, fmt.Errorf("invalid netmap node")
	}

	if len(values) < 1 {
		return nil, fmt.Errorf("invalid netmap node")
	}

	rawNode, err := values[0].TryBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to get node field: %w", err)
	}

	nodeInfo := netmap.NewNodeInfo()
	if err = nodeInfo.Unmarshal(rawNode); err != nil {
		return nil, fmt.Errorf("can't unmarshal peer info: %w", err)
	}

	return nodeInfo, nil
}

func parseCandidate(st stackitem.Item) (*netmap.NodeInfo, error) {
	values, ok := st.Value().([]stackitem.Item)
	if !ok {
		return nil, fmt.Errorf("invalid netmap node")
	}

	if len(values) != 2 {
		return nil, fmt.Errorf("invalid netmap node")
	}

	node, err := parseNode(values[0])
	if err != nil {
		return nil, fmt.Errorf("failed to get node field: %w", err)
	}

	state, err := getInt64(values[1:2])
	if err != nil {
		return nil, fmt.Errorf("failed to get state field: %w", err)
	}

	switch state {
	case 1:
		node.SetState(netmap.NodeStateOnline)
	case 2:
		node.SetState(netmap.NodeStateOffline)
	default:
		node.SetState(0)
	}

	return node, nil
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

func getArray(st []stackitem.Item) ([]stackitem.Item, error) {
	index := len(st) - 1 // top stack element is last in the array
	arr, err := st[index].Convert(stackitem.ArrayT)
	if err != nil {
		return nil, err
	}
	if _, ok := arr.(stackitem.Null); ok {
		return nil, nil
	}

	iterator, ok := arr.Value().([]stackitem.Item)
	if !ok {
		return nil, fmt.Errorf("bad conversion")
	}
	return iterator, nil
}
