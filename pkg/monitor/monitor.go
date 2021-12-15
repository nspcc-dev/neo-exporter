package monitor

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	nns "github.com/nspcc-dev/neo-go/examples/nft-nd-nns"
	"github.com/nspcc-dev/neo-go/pkg/core/native/nativenames"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm"
	"github.com/nspcc-dev/neo-go/pkg/vm/stackitem"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/locode"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/morphchain"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var errUnknownKeyFormat = errors.New("could not parse private key: expected WIF, hex or path to binary key")

type (
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
		FetchNEP17TotalSupply(util.Uint160) (int64, error)
	}

	AlphabetFetcher interface {
		FetchAlphabet() (keys.PublicKeys, error)
	}

	Monitor struct {
		balance util.Uint160
		proxy   *util.Uint160
		neofs   *util.Uint160

		logger                 *zap.Logger
		sleep                  time.Duration
		metricsServer          http.Server
		geoFetcher             *locode.DB
		alpFetcher             AlphabetFetcher
		nmFetcher              NetmapFetcher
		irFetcher              InnerRingFetcher
		sideBlFetcher          BalanceFetcher
		mainBlFetcher          BalanceFetcher
		sideChainNotaryEnabled bool
	}
)

func New(ctx context.Context, cfg *viper.Viper) (*Monitor, error) {
	logConf := zap.NewProductionConfig()
	logConf.Level = WithLevel(cfg.GetString(cfgLoggerLevel))
	logger, err := logConf.Build()
	if err != nil {
		return nil, err
	}

	key, err := readKey(logger, cfg)
	if err != nil {
		return nil, err
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
		Logger:         logger,
	})
	if err != nil {
		return nil, fmt.Errorf("can't initialize netmap fetcher: %w", err)
	}

	alphabetFetcher, err := morphchain.NewAlphabetFetcher(morphchain.AlphabetFetcherArgs{
		Committeer: sideNeogoClient,
	})
	if err != nil {
		return nil, fmt.Errorf("can't initialize alphabet fetcher: %w", err)
	}

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

	var (
		balance util.Uint160
		proxy   *util.Uint160
		neofs   *util.Uint160
	)

	balance, err = getScriptHash(cfg, sideNeogoClient, "balance.neofs", cfgBalanceContract)
	if err != nil {
		return nil, fmt.Errorf("balance contract is not available: %w", err)
	}

	proxyContract, err := getScriptHash(cfg, sideNeogoClient, "proxy.neofs", cfgProxyContract)
	if err != nil {
		logger.Info("proxy disabled")
	} else {
		proxy = &proxyContract
	}

	neofsContract := cfg.GetString(cfgNeoFSContract)
	if len(neofsContract) != 0 {
		sh, err := util.Uint160DecodeStringLE(neofsContract)
		if err != nil {
			return nil, fmt.Errorf("NNS u160 decode: %w", err)
		}
		neofs = &sh
	} else {
		logger.Info("neofs contract ignored")
	}

	_, err = sideNeogoClient.GetNativeContractHash(nativenames.Notary)
	notaryEnabled := err == nil

	geoFetcher := locode.New(
		locode.Prm{
			Path: cfg.GetString(cfgLocodeDB),
		},
	)

	return &Monitor{
		balance: balance,
		proxy:   proxy,
		neofs:   neofs,
		logger:  logger,
		sleep:   cfg.GetDuration(cfgMetricsInterval),
		metricsServer: http.Server{
			Addr:    cfg.GetString(cfgMetricsEndpoint),
			Handler: promhttp.Handler(),
		},
		geoFetcher:             geoFetcher,
		alpFetcher:             alphabetFetcher,
		nmFetcher:              nmFetcher,
		irFetcher:              nmFetcher,
		sideBlFetcher:          sideBalanceFetcher,
		mainBlFetcher:          mainBalanceFetcher,
		sideChainNotaryEnabled: notaryEnabled,
	}, nil
}

func (m *Monitor) Start(ctx context.Context) {
	prometheus.MustRegister(locationPresent)
	prometheus.MustRegister(droppedNodesCount)
	prometheus.MustRegister(newNodesCount)
	prometheus.MustRegister(epochNumber)
	prometheus.MustRegister(storageNodeGASBalances)
	prometheus.MustRegister(storageNodeNotaryBalances)
	prometheus.MustRegister(innerRingBalances)
	prometheus.MustRegister(alphabetGASBalances)
	prometheus.MustRegister(alphabetNotaryBalances)
	prometheus.MustRegister(proxyBalance)
	prometheus.MustRegister(mainChainSupply)
	prometheus.MustRegister(sideChainSupply)

	if err := m.geoFetcher.Open(); err != nil {
		m.logger.Warn("geoposition fetching disabled", zap.Error(err))
	}

	go func() {
		err := m.metricsServer.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			m.logger.Error("start metrics server error", zap.Error(err))
		}
	}()

	go m.Job(ctx)
}

func (m *Monitor) Stop() {
	err := m.metricsServer.Close()
	if err != nil {
		m.logger.Error("stop metrics server error", zap.Error(err))
	}

	_ = m.geoFetcher.Close()
}

func (m *Monitor) Job(ctx context.Context) {
	for {
		m.logger.Debug("scraping data from side chain")

		netmap, err := m.nmFetcher.FetchNetmap()
		if err != nil {
			m.logger.Warn("can't scrap network map info", zap.Error(err))
		} else {
			candidatesNetmap, err := m.nmFetcher.FetchCandidates()
			if err != nil {
				m.logger.Warn("can't scrap network map candidates info", zap.Error(err))
			} else {
				m.processNetworkMap(netmap, candidatesNetmap)
			}
		}

		innerRing, err := m.irFetcher.FetchInnerRingKeys()
		if err != nil {
			m.logger.Warn("can't scrap inner ring info", zap.Error(err))
		} else {
			m.processInnerRing(innerRing)
		}

		if m.proxy != nil {
			m.processProxyContract()
		}

		m.processSideChainSupply()

		if m.neofs != nil {
			m.processMainChainSupply()
		}

		alphabet, err := m.alpFetcher.FetchAlphabet()
		if err != nil {
			m.logger.Warn("can't scrap alphabet info", zap.Error(err))
		} else {
			m.processAlphabet(alphabet)
		}

		select {
		case <-time.After(m.sleep):
			// sleep for some time before next prometheus update
		case <-ctx.Done():
			m.logger.Info("context closed, stop monitor")
			return
		}
	}
}

func (m *Monitor) Logger() *zap.Logger {
	return m.logger
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
		sh, err = nnsResolve(cli, cs.Hash, nnsKey)
		if err != nil {
			return sh, fmt.Errorf("NNS.resolve: %w", err)
		}
	} else {
		sh, err = util.Uint160DecodeStringLE(hash)
		if err != nil {
			return sh, fmt.Errorf("NNS u160 decode: %w", err)
		}
	}

	return sh, nil
}

func nnsResolve(c *client.Client, nnsHash util.Uint160, domain string) (util.Uint160, error) {
	result, err := c.InvokeFunction(nnsHash, "resolve", []smartcontract.Parameter{
		{
			Type:  smartcontract.StringType,
			Value: domain,
		},
		{
			Type:  smartcontract.IntegerType,
			Value: int64(nns.TXT),
		},
	}, nil)
	if err != nil {
		return util.Uint160{}, err
	}
	if result.State != vm.HaltState.String() {
		return util.Uint160{}, fmt.Errorf("invocation failed: %s", result.FaultException)
	}
	if len(result.Stack) == 0 {
		return util.Uint160{}, errors.New("result stack is empty")
	}

	// Parse the result of resolving NNS record.
	// It works with multiple formats (corresponding to multiple NNS versions).
	// If array of hashes is provided, it returns only the first one.
	res := result.Stack[0]
	if arr, ok := res.Value().([]stackitem.Item); ok {
		if len(arr) == 0 {
			return util.Uint160{}, errors.New("NNS record is missing")
		}
		res = arr[0]
	}
	bs, err := res.TryBytes()
	if err != nil {
		return util.Uint160{}, fmt.Errorf("malformed response: %w", err)
	}
	return util.Uint160DecodeStringLE(string(bs))
}

type diffNode struct {
	currEpoch *morphchain.Node
	nextEpoch *morphchain.Node
}

type nodeLocation struct {
	name string
	long string
	lat  string
}

func (m *Monitor) processNetworkMap(nm morphchain.NetmapInfo, candidates morphchain.NetmapCandidatesInfo) {
	currentNetmapLen := len(nm.Nodes)

	exportCountries := make(map[nodeLocation]int, currentNetmapLen)
	exportBalancesGAS := make(map[string]int64, currentNetmapLen)
	exportBalancesNotary := make(map[string]int64, currentNetmapLen)

	newNodes, droppedNodes := getDiff(nm, candidates)

	for _, node := range nm.Nodes {
		keyHex := hex.EncodeToString(node.PublicKey.Bytes())

		balanceGAS, err := m.sideBlFetcher.FetchGAS(*node.PublicKey)
		if err != nil {
			m.logger.Debug("can't fetch GAS balance", zap.String("key", keyHex), zap.Error(err))
		} else {
			exportBalancesGAS[keyHex] = balanceGAS
		}

		pos, err := m.geoFetcher.Get(node)
		if err != nil {
			m.logger.Debug("can't fetch geoposition", zap.String("key", keyHex), zap.Error(err))
		} else {
			nodeLoc := nodeLocation{
				name: pos.Location(),
				long: strconv.FormatFloat(pos.Longitude(), 'f', 4, 64),
				lat:  strconv.FormatFloat(pos.Latitude(), 'f', 4, 64),
			}

			exportCountries[nodeLoc]++
		}

		if m.sideChainNotaryEnabled {
			balanceNotary, err := m.sideBlFetcher.FetchNotary(*node.PublicKey)
			if err != nil {
				m.logger.Debug("can't fetch notary balance", zap.String("key", keyHex), zap.Error(err))
			} else {
				exportBalancesNotary[keyHex] = balanceNotary
			}
		}
	}

	m.logNodes("new node", newNodes)
	m.logNodes("dropped node", droppedNodes)

	epochNumber.Set(float64(nm.Epoch))
	droppedNodesCount.Set(float64(len(droppedNodes)))
	newNodesCount.Set(float64(len(newNodes)))

	locationPresent.Reset()
	for k, v := range exportCountries {
		locationPresent.With(prometheus.Labels{
			location:  k.name,
			longitude: k.long,
			latitude:  k.lat,
		}).Set(float64(v))
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

func (m *Monitor) logNodes(msg string, nodes []*morphchain.Node) {
	for _, node := range nodes {
		fields := []zap.Field{zap.Uint64("id", node.ID), zap.String("address", node.Address),
			zap.String("public key", node.PublicKey.String()),
		}

		for key, val := range node.Attributes {
			fields = append(fields, zap.String(key, val))
		}

		m.logger.Info(msg, fields...)
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
			m.logger.Debug("can't fetch balance", zap.String("key", keyHex), zap.Error(err))
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
		m.logger.Debug("can't fetch proxy contract balance", zap.Error(err))
		return
	}

	proxyBalance.Set(float64(balance))
}

func (m *Monitor) processAlphabet(alphabet keys.PublicKeys) {
	exportGasBalances := make(map[string]int64, len(alphabet))
	exportNotaryBalances := make(map[string]int64, len(alphabet))

	for _, key := range alphabet {
		keyHex := hex.EncodeToString(key.Bytes())

		balanceGAS, err := m.mainBlFetcher.FetchGAS(*key)
		if err != nil {
			m.logger.Debug("can't fetch gas balance", zap.String("key", keyHex), zap.Error(err))
		} else {
			exportGasBalances[keyHex] = balanceGAS
		}

		if m.sideChainNotaryEnabled {
			balanceNotary, err := m.sideBlFetcher.FetchNotary(*key)
			if err != nil {
				m.logger.Debug("can't fetch notary balance", zap.String("key", keyHex), zap.Error(err))
			} else {
				exportNotaryBalances[keyHex] = balanceNotary
			}
		}
	}

	alphabetGASBalances.Reset()
	for k, v := range exportGasBalances {
		alphabetGASBalances.WithLabelValues(k).Set(float64(v))
	}

	alphabetNotaryBalances.Reset()
	for k, v := range exportNotaryBalances {
		alphabetNotaryBalances.WithLabelValues(k).Set(float64(v))
	}
}

func (m *Monitor) processMainChainSupply() {
	balance, err := m.mainBlFetcher.FetchGASByScriptHash(*m.neofs)
	if err != nil {
		m.logger.Debug("can't fetch neofs contract balance", zap.Error(err))
		return
	}

	mainChainSupply.Set(float64(balance))
}

func (m *Monitor) processSideChainSupply() {
	balance, err := m.sideBlFetcher.FetchNEP17TotalSupply(m.balance)
	if err != nil {
		m.logger.Debug("can't fetch balance contract total supply", zap.Error(err))
		return
	}

	sideChainSupply.Set(float64(balance))
}

func readKey(logger *zap.Logger, cfg *viper.Viper) (*keys.PrivateKey, error) {
	var (
		key *keys.PrivateKey
		err error
	)

	keyFromCfg := cfg.GetString(cfgKey)

	if keyFromCfg == "" {
		logger.Debug("using random private key")

		key, err = keys.NewPrivateKey()
		if err != nil {
			return nil, fmt.Errorf("can't generate private key: %w", err)
		}

		return key, nil
	}

	// WIF
	if key, err = keys.NewPrivateKeyFromWIF(keyFromCfg); err == nil {
		logger.Debug("using private key from WIF")
		return key, nil
	}

	// hex
	if key, err = keys.NewPrivateKeyFromHex(keyFromCfg); err == nil {
		logger.Debug("using private key from hex")
		return key, nil
	}

	var data []byte

	// file
	if data, err = os.ReadFile(keyFromCfg); err == nil {
		logger.Debug("using private key from file")
		return keys.NewPrivateKeyFromBytes(data)
	}

	return nil, errUnknownKeyFormat
}
