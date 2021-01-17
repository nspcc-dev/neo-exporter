package morphchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-node/pkg/morph/client"
	morphNetmap "github.com/nspcc-dev/neofs-node/pkg/morph/client/netmap"
	wrapNetmap "github.com/nspcc-dev/neofs-node/pkg/morph/client/netmap/wrapper"
	"github.com/nspcc-dev/neofs-node/pkg/network"
)

type (
	NetmapFetcher struct {
		cli *wrapNetmap.Wrapper
	}

	NetmapFetcherArgs struct {
		Key            *ecdsa.PrivateKey
		Endpoint       string
		DialTimeout    time.Duration
		NetmapContract util.Uint160
	}

	NetmapInfo struct {
		Epoch     uint64
		Addresses []string
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

	staticClient, err := client.NewStatic(
		blockchainClient,
		p.NetmapContract,
		0)
	if err != nil {
		return nil, fmt.Errorf("can't create netmap contract static client: %w", err)
	}

	enhancedNetmapClient, err := morphNetmap.New(staticClient)
	if err != nil {
		return nil, fmt.Errorf("can't create netmap morph client: %w", err)
	}

	wrapper, err := wrapNetmap.New(enhancedNetmapClient)
	if err != nil {
		return nil, fmt.Errorf("can't create netmap client wrapper: %w", err)
	}

	return &NetmapFetcher{
		cli: wrapper,
	}, nil
}

func (f *NetmapFetcher) Fetch() (NetmapInfo, error) {
	epoch, err := f.cli.Epoch()
	if err != nil {
		return NetmapInfo{}, fmt.Errorf("can't fetch epoch number: %w", err)
	}

	nm, err := f.cli.GetNetMap(0)
	if err != nil {
		return NetmapInfo{}, fmt.Errorf("can't fetch network map: %w", err)
	}

	addresses := make([]string, 0, len(nm.Nodes))

	for _, node := range nm.Nodes {
		addr, err := multiAddrToIPStringWithoutPort(node.Address())
		if err != nil {
			log.Printf("morphchain: %s", err.Error())
		}

		addresses = append(addresses, addr)
	}

	return NetmapInfo{
		Epoch:     epoch,
		Addresses: addresses,
	}, nil
}

func multiAddrToIPStringWithoutPort(multiaddr string) (string, error) {
	ipWithPort, err := network.IPAddrFromMultiaddr(multiaddr)
	if err != nil {
		return "", fmt.Errorf("can't transform multiaddr string to ip string: %w", err)
	}

	return strings.Split(ipWithPort, ":")[0], nil
}
