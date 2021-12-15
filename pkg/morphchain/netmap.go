package morphchain

import (
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	neogo "github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-api-go/pkg/netmap"
	morph "github.com/nspcc-dev/neofs-node/pkg/morph/client"
	wrapNetmap "github.com/nspcc-dev/neofs-node/pkg/morph/client/netmap/wrapper"
	"github.com/nspcc-dev/neofs-node/pkg/network"
	"go.uber.org/zap"
)

type (
	NetmapFetcher struct {
		cli            *morph.Client
		notaryDisabled bool
		wrp            *wrapNetmap.Wrapper
		logger         *zap.Logger
	}

	NetmapFetcherArgs struct {
		Cli            *neogo.Client
		Key            *keys.PrivateKey
		NetmapContract util.Uint160
		Logger         *zap.Logger
	}

	Node struct {
		ID         uint64
		Address    string
		PublicKey  *keys.PublicKey
		Attributes map[string]string
	}

	NetmapInfo struct {
		Epoch uint64
		Nodes []*Node
	}

	NetmapCandidatesInfo struct {
		Nodes []*Node
	}
)

func NewNetmapFetcher(p NetmapFetcherArgs) (*NetmapFetcher, error) {
	blockchainClient, err := morph.New(
		p.Key,
		"", // endpoint is ignored due to single client instance
		morph.WithSingleClient(p.Cli),
	)
	if err != nil {
		return nil, fmt.Errorf("can't create blockchain client: %w", err)
	}

	wrapper, err := wrapNetmap.NewFromMorph(
		blockchainClient,
		p.NetmapContract,
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("can't create netmap client wrapper: %w", err)
	}

	return &NetmapFetcher{
		cli:            blockchainClient,
		notaryDisabled: !blockchainClient.ProbeNotary(),
		wrp:            wrapper,
		logger:         p.Logger,
	}, nil
}

func (f *NetmapFetcher) FetchNetmap() (NetmapInfo, error) {
	epoch, err := f.wrp.Epoch()
	if err != nil {
		return NetmapInfo{}, fmt.Errorf("can't fetch epoch number: %w", err)
	}

	nm, err := f.wrp.GetNetMap(0)
	if err != nil {
		return NetmapInfo{}, fmt.Errorf("can't fetch network map: %w", err)
	}

	nodes := make([]*Node, 0, len(nm.Nodes))
	var node *Node

	for _, apiNode := range nm.Nodes {
		node, err = processNode(f.logger, apiNode)
		if err != nil {
			return NetmapInfo{}, err
		}

		nodes = append(nodes, node)
	}

	return NetmapInfo{
		Epoch: epoch,
		Nodes: nodes,
	}, nil
}

func (f *NetmapFetcher) FetchCandidates() (NetmapCandidatesInfo, error) {
	candidatesNetmap, err := f.wrp.GetCandidates()
	if err != nil {
		return NetmapCandidatesInfo{}, fmt.Errorf("can't fetch netmap candidates: %w", err)
	}

	candidates := make([]*Node, 0, len(candidatesNetmap.Nodes))
	var candidate *Node

	for _, apiCandidate := range candidatesNetmap.Nodes {
		candidate, err = processNode(f.logger, apiCandidate)
		if err != nil {
			return NetmapCandidatesInfo{}, nil
		}

		candidates = append(candidates, candidate)
	}

	return NetmapCandidatesInfo{
		Nodes: candidates,
	}, nil
}

func (f *NetmapFetcher) FetchInnerRingKeys() (keys.PublicKeys, error) {
	var (
		publicKeys keys.PublicKeys
		err        error
	)

	if f.notaryDisabled {
		publicKeys, err = f.wrp.GetInnerRingList()
	} else {
		publicKeys, err = f.cli.NeoFSAlphabetList()
	}

	if err != nil {
		return nil, fmt.Errorf("can't fetch inner ring keys: %w", err)
	}

	return publicKeys, nil
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

func processNode(logger *zap.Logger, node *netmap.Node) (*Node, error) {
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

	return &Node{
		ID:         node.ID,
		Address:    address,
		PublicKey:  publicKey,
		Attributes: node.AttrMap,
	}, nil
}
