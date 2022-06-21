package contracts

import (
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/nspcc-dev/neo-go/pkg/core/native/noderoles"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-api-go/pkg/netmap"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/monitor"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/pool"
	"github.com/nspcc-dev/neofs-node/pkg/network"
	"go.uber.org/zap"
)

type (
	Netmap struct {
		pool           *pool.Pool
		contractHash   util.Uint160
		logger         *zap.Logger
		notaryDisabled bool
	}

	NetmapArgs struct {
		Pool           *pool.Pool
		NetmapContract util.Uint160
		Logger         *zap.Logger
	}
)

const (
	epochMethod            = "epoch"
	netmapMethod           = "netmap"
	netmapCandidatesMethod = "netmapCandidates"
	innerRingListMethod    = "innerRingList"
)

// NewNetmap creates Netmap to interact with 'netmap' contract in morph chain.
func NewNetmap(p NetmapArgs) (*Netmap, error) {
	return &Netmap{
		pool:           p.Pool,
		contractHash:   p.NetmapContract,
		notaryDisabled: !p.Pool.ProbeNotary(),
		logger:         p.Logger,
	}, nil
}

func (c *Netmap) FetchNetmap() (monitor.NetmapInfo, error) {
	epoch, err := c.Epoch()
	if err != nil {
		return monitor.NetmapInfo{}, fmt.Errorf("can't fetch epoch number: %w", err)
	}

	apiNodes, err := c.Netmap()
	if err != nil {
		return monitor.NetmapInfo{}, fmt.Errorf("can't fetch network map: %w", err)
	}

	nodes := make([]*monitor.Node, 0, len(apiNodes))
	var node *monitor.Node

	for _, apiNode := range apiNodes {
		node, err = processNode(c.logger, apiNode)
		if err != nil {
			return monitor.NetmapInfo{}, err
		}

		nodes = append(nodes, node)
	}

	return monitor.NetmapInfo{
		Epoch: uint64(epoch),
		Nodes: nodes,
	}, nil
}

func (c *Netmap) FetchCandidates() (monitor.NetmapCandidatesInfo, error) {
	apiCandidates, err := c.NetmapCandidates()
	if err != nil {
		return monitor.NetmapCandidatesInfo{}, fmt.Errorf("can't fetch netmap candidates: %w", err)
	}

	candidates := make([]*monitor.Node, 0, len(apiCandidates))

	for _, apiCandidate := range apiCandidates {
		candidate, err := processNode(c.logger, apiCandidate)
		if err != nil {
			return monitor.NetmapCandidatesInfo{}, nil
		}

		candidates = append(candidates, candidate)
	}

	return monitor.NetmapCandidatesInfo{
		Nodes: candidates,
	}, nil
}

func (c *Netmap) FetchInnerRingKeys() (keys.PublicKeys, error) {
	var (
		publicKeys keys.PublicKeys
		err        error
		height     uint32
	)

	if c.notaryDisabled {
		publicKeys, err = c.InnerRingList()
	} else {
		height, err = c.pool.GetBlockCount()
		if err == nil {
			publicKeys, err = c.pool.GetDesignatedByRole(noderoles.NeoFSAlphabet, height)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("can't fetch inner ring keys: %w", err)
	}

	return publicKeys, nil
}

func (c *Netmap) Epoch() (int64, error) {
	res, err := c.pool.InvokeFunction(c.contractHash, epochMethod, []smartcontract.Parameter{}, nil)
	if err != nil {
		return 0, err
	}

	if err = getInvocationError(res); err != nil {
		return 0, err
	}

	return getInt64(res.Stack)
}

func (c *Netmap) Netmap() ([]*netmap.Node, error) {
	res, err := c.pool.InvokeFunction(c.contractHash, netmapMethod, []smartcontract.Parameter{}, nil)
	if err != nil {
		return nil, err
	}

	if err = getInvocationError(res); err != nil {
		return nil, err
	}

	arr, err := getArray(res.Stack)
	if err != nil {
		return nil, err
	}

	infos := make([]netmap.NodeInfo, 0, len(arr))
	for _, item := range arr {
		nodeInfo, err := parseNode(item)
		if err != nil {
			return nil, err
		}
		infos = append(infos, *nodeInfo)
	}

	return netmap.NodesFromInfo(infos), nil
}

func (c *Netmap) NetmapCandidates() (netmap.Nodes, error) {
	res, err := c.pool.InvokeFunction(c.contractHash, netmapCandidatesMethod, []smartcontract.Parameter{}, nil)
	if err != nil {
		return nil, err
	}

	if err = getInvocationError(res); err != nil {
		return nil, err
	}

	arr, err := getArray(res.Stack)
	if err != nil {
		return nil, err
	}

	candidates := make([]netmap.NodeInfo, 0, len(arr))

	for _, item := range arr {
		nodeInfo, err := parseCandidate(item)
		if err != nil {
			return nil, err
		}

		candidates = append(candidates, *nodeInfo)
	}

	return netmap.NodesFromInfo(candidates), nil
}

func (c *Netmap) InnerRingList() (keys.PublicKeys, error) {
	res, err := c.pool.InvokeFunction(c.contractHash, innerRingListMethod, []smartcontract.Parameter{}, nil)
	if err != nil {
		return nil, err
	}

	if err = getInvocationError(res); err != nil {
		return nil, err
	}

	arr, err := getArray(res.Stack)
	if err != nil {
		return nil, err
	}

	irKeys := make(keys.PublicKeys, 0, len(arr))

	for _, item := range arr {
		irKey, err := parseIRNode(item)
		if err != nil {
			return nil, err
		}
		irKeys = append(irKeys, irKey)
	}

	return irKeys, nil
}

func processNode(logger *zap.Logger, node *netmap.Node) (*monitor.Node, error) {
	var address string

	node.IterateAddresses(
		func(mAddr string) bool {
			addr, err := multiAddrToIPStringWithoutPort(mAddr)
			if err != nil {
				logger.Debug("morphchain", zap.Error(err))
				return false
			}

			// TODO: define if monitor should show
			//  all addresses of the node or only
			//  one of them: #17.
			address = addr

			return true
		},
	)

	rawPublicKey := node.PublicKey()

	publicKey, err := keys.NewPublicKeyFromBytes(rawPublicKey, elliptic.P256())
	if err != nil {
		return nil, fmt.Errorf(
			"can't parse storage node public key <%s>: %w",
			hex.EncodeToString(rawPublicKey),
			err,
		)
	}

	return &monitor.Node{
		ID:         node.ID,
		Address:    address,
		PublicKey:  publicKey,
		Attributes: node.AttrMap,
	}, nil
}

func multiAddrToIPStringWithoutPort(multiaddr string) (string, error) {
	var netAddress network.Address

	err := netAddress.FromString(multiaddr)
	if err != nil {
		return "", err
	}

	ipWithPort := netAddress.HostAddr()

	return strings.Split(ipWithPort, ":")[0], nil
}
