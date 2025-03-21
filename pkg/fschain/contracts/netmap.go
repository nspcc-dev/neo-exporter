package contracts

import (
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"

	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/nspcc-dev/hrw/v2"
	"github.com/nspcc-dev/neo-exporter/pkg/monitor"
	"github.com/nspcc-dev/neo-exporter/pkg/pool"
	"github.com/nspcc-dev/neo-go/pkg/core/native/noderoles"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/util"
	rpcnetmap "github.com/nspcc-dev/neofs-contract/rpc/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"go.uber.org/zap"
)

type (
	Netmap struct {
		pool   *pool.Pool
		logger *zap.Logger

		contractReader *rpcnetmap.ContractReader
	}

	NetmapArgs struct {
		Pool           *pool.Pool
		NetmapContract util.Uint160
		Logger         *zap.Logger
	}
)

const (
	grpcScheme    = "grpc"
	grpcTLSScheme = "grpcs"
)

// NewNetmap creates Netmap to interact with 'netmap' contract in FS chain.
func NewNetmap(p NetmapArgs) (*Netmap, error) {
	return &Netmap{
		pool:           p.Pool,
		logger:         p.Logger,
		contractReader: rpcnetmap.NewReader(p.Pool, p.NetmapContract),
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

	height, err = c.pool.GetBlockCount()
	if err == nil {
		publicKeys, err = c.pool.GetDesignatedByRole(noderoles.NeoFSAlphabet, height)
	}

	if err != nil {
		return nil, fmt.Errorf("can't fetch inner ring keys: %w", err)
	}

	return publicKeys, nil
}

func (c *Netmap) Epoch() (int64, error) {
	e, err := c.contractReader.Epoch()
	if err != nil {
		return 0, fmt.Errorf("epoch: %w", err)
	}

	return e.Int64(), nil
}

func (c *Netmap) Netmap() ([]*netmap.NodeInfo, error) {
	return c.parsedNodes((*rpcnetmap.ContractReader).Netmap)
}

func (c *Netmap) NetmapCandidates() ([]*netmap.NodeInfo, error) {
	return c.parsedNodes((*rpcnetmap.ContractReader).NetmapCandidates)
}

func (c *Netmap) parsedNodes(f func(reader *rpcnetmap.ContractReader) ([]*rpcnetmap.NetmapNode, error)) ([]*netmap.NodeInfo, error) {
	data, err := f(c.contractReader)
	if err != nil {
		return nil, fmt.Errorf("contract reader: %w", err)
	}

	nodes, err := parseContractNodes(data)
	if err != nil {
		return nil, fmt.Errorf("parseContractNodes: %w", err)
	}

	return nodes, nil
}

func processNode(logger *zap.Logger, node *netmap.NodeInfo) (*monitor.Node, error) {
	var address string

	node.IterateNetworkEndpoints(
		func(mAddr string) bool {
			addr, err := multiAddrToIPStringWithoutPort(mAddr)
			if err != nil {
				logger.Debug("FS chain", zap.Error(err))
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

	attrs := make(map[string]string, node.NumberOfAttributes())
	node.IterateAttributes(func(key, value string) {
		attrs[key] = value
	})

	return &monitor.Node{
		ID:         hrw.Hash(node.PublicKey()),
		Address:    address,
		PublicKey:  publicKey,
		Attributes: attrs,
		Locode:     node.LOCODE(),
	}, nil
}

func multiAddrToIPStringWithoutPort(multiaddress string) (string, error) {
	var host string
	if netAddress, err := multiaddr.NewMultiaddr(multiaddress); err != nil {
		if host, _, err = parseURI(multiaddress); err != nil {
			return "", err
		}
	} else if _, host, err = manet.DialArgs(netAddress); err != nil {
		return "", err
	}

	uriAddress := (&url.URL{Scheme: "grpc", Host: host}).String()
	// we need this to splitHostPort
	URL, err := url.Parse(uriAddress)
	if err != nil {
		return "", err
	}

	return URL.Hostname(), nil
}

// ParseURI parses s as address and returns a host and a flag
// indicating that TLS is enabled. If multi-address is provided
// the argument is returned unchanged.
func parseURI(s string) (string, bool, error) {
	uri, err := url.ParseRequestURI(s)
	if err != nil {
		return s, false, nil
	}

	// check if passed string was parsed correctly
	// URIs that do not start with a slash after the scheme are interpreted as:
	// `scheme:opaque` => if `opaque` is not empty, then it is supposed that URI
	// is in `host:port` format
	if uri.Host == "" {
		uri.Host = uri.Scheme
		uri.Scheme = grpcScheme // assume GRPC by default
		if uri.Opaque != "" {
			uri.Host = net.JoinHostPort(uri.Host, uri.Opaque)
		}
	}

	switch uri.Scheme {
	case grpcTLSScheme, grpcScheme:
	default:
		return "", false, fmt.Errorf("unsupported scheme: %s", uri.Scheme)
	}

	return uri.Host, uri.Scheme == grpcTLSScheme, nil
}
