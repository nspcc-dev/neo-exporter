package monitor

import (
	"context"
	"encoding/hex"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/nspcc-dev/locode-db/pkg/locodedb"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type (
	NetmapFetcher interface {
		FetchNetmap() (NetmapInfo, error)
		FetchCandidates() (NetmapCandidatesInfo, error)
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
		FetchSideAlphabet() (keys.PublicKeys, error)
		FetchMainAlphabet() (keys.PublicKeys, error)
	}

	ContainerFetcher interface {
		Total() (int64, error)
	}

	Node struct {
		ID         uint64
		Address    string
		PublicKey  *keys.PublicKey
		Attributes map[string]string
		Locode     string
	}

	NetmapInfo struct {
		Epoch uint64
		Nodes []*Node
	}

	NetmapCandidatesInfo struct {
		Nodes []*Node
	}

	Monitor struct {
		balance util.Uint160
		proxy   *util.Uint160
		neofs   *util.Uint160

		logger        *zap.Logger
		sleep         time.Duration
		metricsServer http.Server
		alpFetcher    AlphabetFetcher
		nmFetcher     NetmapFetcher
		irFetcher     InnerRingFetcher
		sideBlFetcher BalanceFetcher
		mainBlFetcher BalanceFetcher
		cnrFetcher    ContainerFetcher
	}

	Args struct {
		Balance        util.Uint160
		Proxy          *util.Uint160
		Neofs          *util.Uint160
		Logger         *zap.Logger
		Sleep          time.Duration
		MetricsAddress string
		AlpFetcher     AlphabetFetcher
		NmFetcher      NetmapFetcher
		IRFetcher      InnerRingFetcher
		SideBlFetcher  BalanceFetcher
		MainBlFetcher  BalanceFetcher
		CnrFetcher     ContainerFetcher
	}
)

func New(p Args) *Monitor {
	return &Monitor{
		balance: p.Balance,
		proxy:   p.Proxy,
		neofs:   p.Neofs,
		logger:  p.Logger,
		sleep:   p.Sleep,
		metricsServer: http.Server{
			Addr:    p.MetricsAddress,
			Handler: promhttp.Handler(),
		},
		alpFetcher:    p.AlpFetcher,
		nmFetcher:     p.NmFetcher,
		irFetcher:     p.IRFetcher,
		sideBlFetcher: p.SideBlFetcher,
		mainBlFetcher: p.MainBlFetcher,
		cnrFetcher:    p.CnrFetcher,
	}
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
	prometheus.MustRegister(alphabetDivergence)
	prometheus.MustRegister(alphabetMainDivergence)
	prometheus.MustRegister(alphabetSideDivergence)
	prometheus.MustRegister(containersNumber)

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

		if sideAlphabet, err := m.alpFetcher.FetchSideAlphabet(); err != nil {
			m.logger.Warn("can't scrap side alphabet info", zap.Error(err))
		} else {
			m.processAlphabet(sideAlphabet)

			if mainAlphabet, err := m.alpFetcher.FetchMainAlphabet(); err != nil {
				m.logger.Warn("can't scrap main alphabet info", zap.Error(err))
			} else {
				m.processAlphabetDivergence(mainAlphabet, sideAlphabet)
			}
		}

		m.processContainersNumber()

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

type diffNode struct {
	currEpoch *Node
	nextEpoch *Node
}

type nodeLocation struct {
	name string
	long string
	lat  string
}

func (m *Monitor) processNetworkMap(nm NetmapInfo, candidates NetmapCandidatesInfo) {
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

		record, err := locodedb.Get(node.Locode)
		if err != nil {
			m.logger.Debug("can't fetch geoposition", zap.String("key", keyHex), zap.Error(err))
		} else {
			nodeLoc := nodeLocation{
				name: record.Location,
				long: strconv.FormatFloat(record.Point.Longitude(), 'f', 4, 64),
				lat:  strconv.FormatFloat(record.Point.Latitude(), 'f', 4, 64),
			}

			exportCountries[nodeLoc]++
		}

		balanceNotary, err := m.sideBlFetcher.FetchNotary(*node.PublicKey)
		if err != nil {
			m.logger.Debug("can't fetch notary balance", zap.String("key", keyHex), zap.Error(err))
		} else {
			exportBalancesNotary[keyHex] = balanceNotary
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

func (m *Monitor) logNodes(msg string, nodes []*Node) {
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

func getDiff(nm NetmapInfo, cand NetmapCandidatesInfo) ([]*Node, []*Node) {
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

	droppedNodes := make([]*Node, 0, droppedCount)
	newNodes := make([]*Node, 0, newCount)

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

		balanceNotary, err := m.sideBlFetcher.FetchNotary(*key)
		if err != nil {
			m.logger.Debug("can't fetch notary balance", zap.String("key", keyHex), zap.Error(err))
		} else {
			exportNotaryBalances[keyHex] = balanceNotary
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

const (
	mainChainDivergenceLabel = "main"
	sideChainDivergenceLabel = "side"
)

func (m *Monitor) processAlphabetDivergence(mainAlphabet, sideAlphabet keys.PublicKeys) {
	sortedMain := sortedAlphabet(mainAlphabet)
	sortedSide := sortedAlphabet(sideAlphabet)

	uniqueMain, uniqueSide := computeUniqueAlphabets(sortedMain, sortedSide)

	alphabetDivergence.Reset()
	alphabetDivergence.WithLabelValues(mainChainDivergenceLabel).Set(float64(len(uniqueMain)))
	alphabetDivergence.WithLabelValues(sideChainDivergenceLabel).Set(float64(len(uniqueSide)))

	alphabetMainDivergence.Reset()
	for _, key := range uniqueMain {
		alphabetMainDivergence.WithLabelValues(key).Set(1)
	}
	alphabetSideDivergence.Reset()
	for _, key := range uniqueSide {
		alphabetSideDivergence.WithLabelValues(key).Set(1)
	}
}

func computeUniqueAlphabets(sortedMain, sortedSide []string) ([]string, []string) {
	var uniqueMain, uniqueSide []string

	i, j := 0, 0
	len1, len2 := len(sortedMain), len(sortedSide)
	for i < len1 && j < len2 {
		if sortedMain[i] == sortedSide[j] {
			i++
			j++
			continue
		}

		if sortedMain[i] < sortedSide[j] {
			uniqueMain = append(uniqueMain, sortedMain[i])
			i++
		} else {
			uniqueSide = append(uniqueSide, sortedSide[j])
			j++
		}
	}

	if i == len1 {
		uniqueSide = append(uniqueSide, sortedSide[j:]...)
	} else if j == len2 {
		uniqueMain = append(uniqueMain, sortedMain[i:]...)
	}

	return uniqueMain, uniqueSide
}

func sortedAlphabet(alphabet keys.PublicKeys) []string {
	sort.Sort(alphabet)
	sorted := make([]string, 0, len(alphabet))
	for _, key := range alphabet {
		keyHex := hex.EncodeToString(key.Bytes())
		sorted = append(sorted, keyHex)
	}
	return sorted
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

func (m *Monitor) processContainersNumber() {
	total, err := m.cnrFetcher.Total()
	if err != nil {
		m.logger.Warn("can't fetch number of available containers", zap.Error(err))
		return
	}

	containersNumber.Set(float64(total))
}
