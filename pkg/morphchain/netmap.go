package morphchain

import (
	"context"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-node/pkg/morph/client"
	wrapNetmap "github.com/nspcc-dev/neofs-node/pkg/morph/client/netmap/wrapper"
	"github.com/nspcc-dev/neofs-node/pkg/network"
)

type (
	NetmapFetcher struct {
		cli *client.Client
		wrp *wrapNetmap.Wrapper
	}

	NetmapFetcherArgs struct {
		Key            *keys.PrivateKey
		Endpoint       string
		DialTimeout    time.Duration
		NetmapContract util.Uint160
	}

	NetmapInfo struct {
		Epoch      uint64
		Addresses  []string
		PublicKeys keys.PublicKeys
	}
)

func NewNetmapFetcher(ctx context.Context, p NetmapFetcherArgs) (*NetmapFetcher, error) {
	blockchainClient, err := client.New(
		p.Key,
		p.Endpoint,
		client.WithContext(ctx),
		client.WithDialTimeout(p.DialTimeout),
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
		cli: blockchainClient,
		wrp: wrapper,
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

	// TODO: define how to allocate slice
	//  since `len(nm.Nodes)` is not enough
	//  for all groups of addresses
	addresses := make([]string, 0, len(nm.Nodes))
	publicKeys := make(keys.PublicKeys, 0, len(nm.Nodes))

	for _, node := range nm.Nodes {
		node.IterateAddresses(
			func(mAddr string) bool {
				addr, err := multiAddrToIPStringWithoutPort(mAddr)
				if err != nil {
					log.Printf("morphchain: %s", err)
				} else {
					addresses = append(addresses, addr)
				}

				// TODO: define if monitor should show
				//  all addresses of the node or only
				//  one of them
				return false
			},
		)

		rawPublicKey := node.PublicKey()

		publicKey, err := keys.NewPublicKeyFromBytes(rawPublicKey, elliptic.P256())
		if err != nil {
			return NetmapInfo{}, fmt.Errorf("can't parse storage node public key <%s>: %w",
				hex.EncodeToString(rawPublicKey), err)
		} else {
			publicKeys = append(publicKeys, publicKey)
		}
	}

	return NetmapInfo{
		Epoch:      epoch,
		Addresses:  addresses,
		PublicKeys: publicKeys,
	}, nil
}

func (f *NetmapFetcher) FetchInnerRingKeys() (keys.PublicKeys, error) {
	publicKeys, err := f.cli.NeoFSAlphabetList()
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
