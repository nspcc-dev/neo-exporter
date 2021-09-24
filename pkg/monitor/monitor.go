package monitor

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	nns "github.com/nspcc-dev/neo-go/examples/nft-nd-nns"
	"github.com/nspcc-dev/neo-go/pkg/core/native/nativenames"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/geoip"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/morphchain"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

var errUnknownKeyFormat = errors.New("could not parse private key: expected WIF, hex or path to binary key")

type (
	GeoIPFetcher interface {
		Fetch(string) (geoip.Info, error)
	}

	NetmapFetcher interface {
		FetchNetmap() (morphchain.NetmapInfo, error)
		FetchCandidates() (morphchain.NetmapCandidatesInfo, error)
	}

	InnerRingFetcher interface {
		FetchInnerRingKeys() (keys.PublicKeys, error)
	}

	BalanceFetcher interface {
		FetchGAS(keys.PublicKey) (int64, error)
		FetchGASByScriptHash(uint160 util.Uint160) (int64, error)
		FetchNotary(keys.PublicKey) (int64, error)
		FetchNotaryByScriptHash(uint160 util.Uint160) (int64, error)
	}

	AlphabetFetcher interface {
		FetchAlphabet() (keys.PublicKeys, error)
	}

	Monitor struct {
		proxy                  *util.Uint160
		sleep                  time.Duration
		metricsServer          http.Server
		ipFetcher              GeoIPFetcher
		alpFetcher             AlphabetFetcher
		nmFetcher              NetmapFetcher
		irFetcher              InnerRingFetcher
		sideBlFetcher          BalanceFetcher
		mainBlFetcher          BalanceFetcher
		sideChainNotaryEnabled bool
	}
)

func New(ctx context.Context, cfg *viper.Viper) (*Monitor, error) {
	key, err := readKey(cfg)
	if err != nil {
		return nil, err
	}

	ipFetcher, err := geoip.NewCachedFetcher(geoip.FetcherArgs{
		Timeout:   cfg.GetDuration(cfgGeoIPDialTimeout),
		Endpoint:  cfg.GetString(cfgGeoIPEndpoint),
		AccessKey: cfg.GetString(cfgGeoIPAccessKey),
	})
	if err != nil {
		return nil, fmt.Errorf("can't initialize geoip fetcher: %w", err)
	}

	sideChainEndpoint := cfg.GetString(sidePrefix + delimiter + cfgNeoRPCEndpoint)
	sideChainTimeout := cfg.GetDuration(sidePrefix + delimiter + cfgNeoRPCDialTimeout)

	mainChainEndpoint := cfg.GetString(mainPrefix + delimiter + cfgNeoRPCEndpoint)
	mainChainTimeout := cfg.GetDuration(mainPrefix + delimiter + cfgNeoRPCDialTimeout)

	sideNeogoClient, err := neoGoClient(ctx, sideChainEndpoint, client.Options{
		DialTimeout: sideChainTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("can't create side chain neo-go client: %w", err)
	}

	mainNeogoClient, err := neoGoClient(ctx, mainChainEndpoint, client.Options{
		DialTimeout: mainChainTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("can't create main chain neo-go client: %w", err)
	}

	netmapContract, err := getScriptHash(cfg, sideNeogoClient, "netmap.neofs", cfgNetmapContract)
	if err != nil {
		return nil, fmt.Errorf("can't read netmap scripthash: %w", err)
	}

	nmFetcher, err := morphchain.NewNetmapFetcher(morphchain.NetmapFetcherArgs{
		Cli:            sideNeogoClient,
		Key:            key,
		NetmapContract: netmapContract,
	})
	if err != nil {
		return nil, fmt.Errorf("can't initialize netmap fetcher: %w", err)
	}

	alphabetFetcher, err := morphchain.NewAlphabetFetcher(morphchain.AlphabetFetcherArgs{
		Committeer: sideNeogoClient,
	})

	sideBalanceFetcher, err := morphchain.NewBalanceFetcher(morphchain.BalanceFetcherArgs{
		Cli: sideNeogoClient,
	})
	if err != nil {
		return nil, fmt.Errorf("can't initialize side balance fetcher: %w", err)
	}

	mainBalanceFetcher, err := morphchain.NewBalanceFetcher(morphchain.BalanceFetcherArgs{
		Cli: mainNeogoClient,
	})
	if err != nil {
		return nil, fmt.Errorf("can't initialize main balance fetcher: %w", err)
	}

	var proxy *util.Uint160

	proxyContract, err := getScriptHash(cfg, sideNeogoClient, "proxy.neofs", cfgProxyContract)
	if err != nil {
		log.Println("proxy disabled")
	} else {
		proxy = &proxyContract
	}

	_, err = sideNeogoClient.GetNativeContractHash(nativenames.Notary)
	notaryEnabled := err == nil

	return &Monitor{
		proxy: proxy,
		sleep: cfg.GetDuration(cfgMetricsInterval),
		metricsServer: http.Server{
			Addr:    cfg.GetString(cfgMetricsEndpoint),
			Handler: promhttp.Handler(),
		},
		ipFetcher:              ipFetcher,
		alpFetcher:             alphabetFetcher,
		nmFetcher:              nmFetcher,
		irFetcher:              nmFetcher,
		sideBlFetcher:          sideBalanceFetcher,
		mainBlFetcher:          mainBalanceFetcher,
		sideChainNotaryEnabled: notaryEnabled,
	}, nil
}

func (m *Monitor) Start(ctx context.Context) {
	prometheus.MustRegister(countriesPresent)
	prometheus.MustRegister(droppedNodesCount)
	prometheus.MustRegister(newNodesCount)
	prometheus.MustRegister(epochNumber)
	prometheus.MustRegister(storageNodeGASBalances)
	prometheus.MustRegister(storageNodeNotaryBalances)
	prometheus.MustRegister(innerRingBalances)
	prometheus.MustRegister(alphabetBalances)
	prometheus.MustRegister(proxyBalance)

	go func() {
		err := m.metricsServer.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("monitor: strart metrics server error %s", err.Error())
		}
	}()

	go m.Job(ctx)
}

func (m *Monitor) Stop() {
	err := m.metricsServer.Close()
	if err != nil {
		log.Printf("monitor: stop metrics server error %s", err.Error())
	}
}

func (m *Monitor) Job(ctx context.Context) {
	for {
		log.Println("monitor: scraping data from side chain")

		netmap, err := m.nmFetcher.FetchNetmap()
		if err != nil {
			log.Printf("monitor: can't scrap network map info, %s", err)
		} else {
			candidatesNetmap, err := m.nmFetcher.FetchCandidates()
			if err != nil {
				log.Printf("monitor: can't scrap network map candidates info, %s", err)
			} else {
				m.processNetworkMap(netmap, candidatesNetmap)
			}
		}

		innerRing, err := m.irFetcher.FetchInnerRingKeys()
		if err != nil {
			log.Printf("monitor: can't scrap inner ring info, %s", err)
		} else {
			m.processInnerRing(innerRing)
		}

		if m.proxy != nil {
			m.processProxyContract()
		}

		alphabet, err := m.alpFetcher.FetchAlphabet()
		if err != nil {
			log.Printf("monitor: can't scrap alphabet info, %s", err)
		} else {
			m.processAlphabet(alphabet)
		}

		select {
		case <-time.After(m.sleep):
			// sleep for some time before next prometheus update
		case <-ctx.Done():
			log.Println("monitor: context closed, stop monitor")
			return
		}
	}
}

func neoGoClient(ctx context.Context, endpoint string, opts client.Options) (*client.Client, error) {
	cli, err := client.New(ctx, endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("can't create neo-go client: %w", err)
	}

	err = cli.Init()
	if err != nil {
		return nil, fmt.Errorf("can't init neo-go client: %w", err)
	}

	return cli, nil
}

const nnsContractID = 1

func getScriptHash(cfg *viper.Viper, cli *client.Client, nnsKey, configKey string) (sh util.Uint160, err error) {
	cs, err := cli.GetContractStateByID(nnsContractID)
	if err != nil {
		return sh, fmt.Errorf("NNS contract state: %w", err)
	}

	hash := cfg.GetString(configKey)
	if len(hash) == 0 {
		hash, err = cli.NNSResolve(cs.Hash, nnsKey, nns.TXT)
		if err != nil {
			return sh, fmt.Errorf("NNS.resolve: %w", err)
		}
	}

	sh, err = util.Uint160DecodeStringLE(hash)
	if err != nil {
		return sh, fmt.Errorf("NNS u160 decode: %w", err)
	}

	return sh, nil
}

type diffNode struct {
	currEpoch *morphchain.Node
	nextEpoch *morphchain.Node
}

func (m *Monitor) processNetworkMap(nm morphchain.NetmapInfo, candidates morphchain.NetmapCandidatesInfo) {
	currentNetmapLen := len(nm.Nodes)

	exportCountries := make(map[string]int, currentNetmapLen)
	exportBalancesGAS := make(map[string]int64, currentNetmapLen)
	exportBalancesNotary := make(map[string]int64, currentNetmapLen)

	newNodes, droppedNodes := getDiff(nm, candidates)

	for _, node := range nm.Nodes {
		info, err := m.ipFetcher.Fetch(node.Address)
		if err != nil {
			log.Printf("monitor: can't fetch %s info, %s", node.Address, err)
		} else {
			exportCountries[info.CountryCode]++
		}

		keyHex := hex.EncodeToString(node.PublicKey.Bytes())

		balanceGAS, err := m.sideBlFetcher.FetchGAS(*node.PublicKey)
		if err != nil {
			log.Printf("monitor: can't fetch %s GAS balance, %s", keyHex, err)
		} else {
			exportBalancesGAS[keyHex] = balanceGAS
		}

		if m.sideChainNotaryEnabled {
			balanceNotary, err := m.sideBlFetcher.FetchNotary(*node.PublicKey)
			if err != nil {
				log.Printf("monitor: can't fetch %s notary balance, %s", keyHex, err)
			} else {
				exportBalancesNotary[keyHex] = balanceNotary
			}
		}
	}

	epochNumber.Set(float64(nm.Epoch))
	droppedNodesCount.Set(float64(len(droppedNodes)))
	newNodesCount.Set(float64(len(newNodes)))

	countriesPresent.Reset()
	for k, v := range exportCountries {
		countriesPresent.WithLabelValues(k).Set(float64(v))
	}

	storageNodeGASBalances.Reset()
	for k, v := range exportBalancesGAS {
		storageNodeGASBalances.WithLabelValues(k).Set(float64(v))
	}

	storageNodeNotaryBalances.Reset()
	for k, v := range exportBalancesNotary {
		storageNodeNotaryBalances.WithLabelValues(k).Set(float64(v))
	}
}

func getDiff(nm morphchain.NetmapInfo, cand morphchain.NetmapCandidatesInfo) ([]*morphchain.Node, []*morphchain.Node) {
	currentNetmapLen := len(nm.Nodes)
	candidatesLen := len(cand.Nodes)

	diff := make(map[uint64]*diffNode, currentNetmapLen+candidatesLen)

	for _, currEpochNode := range nm.Nodes {
		diff[currEpochNode.ID] = &diffNode{currEpoch: currEpochNode}
	}

	var newCount int

	for _, nextEpochNode := range cand.Nodes {
		if _, exists := diff[nextEpochNode.ID]; exists {
			diff[nextEpochNode.ID].nextEpoch = nextEpochNode
		} else {
			newCount++
			diff[nextEpochNode.ID] = &diffNode{nextEpoch: nextEpochNode}
		}
	}

	droppedCount := currentNetmapLen - (candidatesLen - newCount)

	droppedNodes := make([]*morphchain.Node, 0, droppedCount)
	newNodes := make([]*morphchain.Node, 0, newCount)

	for _, node := range diff {
		if node.nextEpoch == nil {
			droppedNodes = append(droppedNodes, node.currEpoch)
		}

		if node.currEpoch == nil {
			newNodes = append(newNodes, node.nextEpoch)
		}
	}

	return newNodes, droppedNodes
}

func (m *Monitor) processInnerRing(ir keys.PublicKeys) {
	exportBalances := make(map[string]int64, len(ir))

	for _, key := range ir {
		keyHex := hex.EncodeToString(key.Bytes())

		balance, err := m.sideBlFetcher.FetchGAS(*key)
		if err != nil {
			log.Printf("monitor: can't fetch %s balance, %s", keyHex, err)
			continue
		}

		exportBalances[keyHex] = balance
	}

	innerRingBalances.Reset()
	for k, v := range exportBalances {
		innerRingBalances.WithLabelValues(k).Set(float64(v))
	}
}

func (m *Monitor) processProxyContract() {
	balance, err := m.sideBlFetcher.FetchGASByScriptHash(*m.proxy)
	if err != nil {
		log.Printf("monitor: can't fetch proxy contract balance, %s", err)
		return
	}

	proxyBalance.Set(float64(balance))
}

func (m *Monitor) processAlphabet(alphabet keys.PublicKeys) {
	exportBalances := make(map[string]int64, len(alphabet))

	for _, key := range alphabet {
		keyHex := hex.EncodeToString(key.Bytes())

		balance, err := m.mainBlFetcher.FetchGAS(*key)
		if err != nil {
			log.Printf("monitor: can't fetch %s balance, %s", keyHex, err)
			continue
		}

		exportBalances[keyHex] = balance
	}

	alphabetBalances.Reset()
	for k, v := range exportBalances {
		alphabetBalances.WithLabelValues(k).Set(float64(v))
	}
}

func readKey(cfg *viper.Viper) (*keys.PrivateKey, error) {
	var (
		key *keys.PrivateKey
		err error
	)

	keyFromCfg := cfg.GetString(cfgKey)

	if keyFromCfg == "" {
		log.Println("monitor: using random private key")

		key, err = keys.NewPrivateKey()
		if err != nil {
			return nil, fmt.Errorf("monitor: can't generate private key: %w", err)
		}

		return key, nil
	}

	// WIF
	if key, err = keys.NewPrivateKeyFromWIF(keyFromCfg); err == nil {
		log.Println("monitor: using private key from WIF")
		return key, nil
	}

	// hex
	if key, err = keys.NewPrivateKeyFromHex(keyFromCfg); err == nil {
		log.Println("monitor: using private key from hex")
		return key, nil
	}

	var data []byte

	// file
	if data, err = os.ReadFile(keyFromCfg); err == nil {
		log.Println("monitor: using private key from file")
		return keys.NewPrivateKeyFromBytes(data)
	}

	return nil, errUnknownKeyFormat
}
