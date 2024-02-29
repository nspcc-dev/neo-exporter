package monitor

import (
	"encoding/hex"
	"strconv"

	"github.com/nspcc-dev/locode-db/pkg/locodedb"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/gas"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type (
	SideJobArgs struct {
		Logger               *zap.Logger
		Balance              util.Uint160
		Proxy                *util.Uint160
		AlphabetFetcher      AlphabetFetcher
		NmFetcher            NetmapFetcher
		IRFetcher            InnerRingFetcher
		BalanceFetcher       Nep17BalanceFetcher
		NotaryBalanceFetcher NotaryBalanceFetcher
		CnrFetcher           ContainerFetcher
		HeightFetcher        HeightFetcher
		StateFetcher         StateFetcher
		Nep17tracker         *Nep17tracker
	}

	SideJob struct {
		logger               *zap.Logger
		nmFetcher            NetmapFetcher
		irFetcher            InnerRingFetcher
		balanceFetcher       Nep17BalanceFetcher
		notaryBalanceFetcher NotaryBalanceFetcher
		proxy                *util.Uint160
		cnrFetcher           ContainerFetcher
		heightFetcher        HeightFetcher
		stateFetcher         StateFetcher
		alphabetFetcher      AlphabetFetcher
		balance              util.Uint160
		nep17tracker         *Nep17tracker
	}

	diffNode struct {
		currEpoch *Node
		nextEpoch *Node
	}

	nodeLocation struct {
		name string
		long string
		lat  string
	}

	NetmapCandidatesInfo struct {
		Nodes []*Node
	}

	ContainerFetcher interface {
		Total() (int64, error)
	}

	HeightFetcher interface {
		FetchHeight() []HeightData
	}

	StateFetcher interface {
		FetchState(height uint32) []StateData
	}

	HeightData struct {
		Host  string
		Value uint32
	}

	StateData struct {
		Host  string
		Value string
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

	NetmapFetcher interface {
		FetchNetmap() (NetmapInfo, error)
		FetchCandidates() (NetmapCandidatesInfo, error)
	}

	InnerRingFetcher interface {
		FetchInnerRingKeys() (keys.PublicKeys, error)
	}
)

func NewSideJob(args SideJobArgs) *SideJob {
	return &SideJob{
		logger:               args.Logger,
		nmFetcher:            args.NmFetcher,
		irFetcher:            args.IRFetcher,
		balanceFetcher:       args.BalanceFetcher,
		notaryBalanceFetcher: args.NotaryBalanceFetcher,
		proxy:                args.Proxy,
		cnrFetcher:           args.CnrFetcher,
		heightFetcher:        args.HeightFetcher,
		stateFetcher:         args.StateFetcher,
		alphabetFetcher:      args.AlphabetFetcher,
		balance:              args.Balance,
		nep17tracker:         args.Nep17tracker,
	}
}

func (m *SideJob) Process() {
	m.logger.Debug("retrieving data from side chain")

	netmap, err := m.nmFetcher.FetchNetmap()
	if err != nil {
		m.logger.Warn("can't read NeoFS network map", zap.Error(err))
	} else {
		candidatesNetmap, err := m.nmFetcher.FetchCandidates()
		if err != nil {
			m.logger.Warn("can't read NeoFS network map candidates", zap.Error(err))
		} else {
			m.processNetworkMap(netmap, candidatesNetmap)
		}
	}

	innerRing, err := m.irFetcher.FetchInnerRingKeys()
	if err != nil {
		m.logger.Warn("can't read NeoFS Inner Ring members", zap.Error(err))
	} else {
		m.processInnerRing(innerRing)
	}

	if m.proxy != nil {
		m.processProxyContract()
	}

	m.processSideChainSupply()

	if alphabet, err := m.alphabetFetcher.FetchAlphabet(); err != nil {
		m.logger.Warn("can't read NeoFS ALphabet members", zap.Error(err))
	} else {
		processAlphabetPublicKeys(alphabet)
		m.processSideAlphabet(alphabet)
	}

	m.processContainersNumber()

	minHeight := m.processChainHeight()
	m.processChainState(minHeight)
	m.processNep17tracker()
}

func (m *SideJob) processNep17tracker() {
	if m.nep17tracker != nil {
		m.nep17tracker.Process(nep17tracker, nep17trackerTotal)
	}
}

func (m *SideJob) processNetworkMap(nm NetmapInfo, candidates NetmapCandidatesInfo) {
	currentNetmapLen := len(nm.Nodes)

	exportCountries := make(map[nodeLocation]int, currentNetmapLen)
	exportBalancesGAS := make(map[string]float64, currentNetmapLen)
	exportBalancesNotary := make(map[string]float64, currentNetmapLen)

	newNodes, droppedNodes := getDiff(nm, candidates)

	for _, node := range nm.Nodes {
		keyHex := hex.EncodeToString(node.PublicKey.Bytes())
		scriptHash := node.PublicKey.GetScriptHash()

		balanceGAS, err := m.balanceFetcher.Fetch(gas.Hash, scriptHash)
		if err != nil {
			m.logger.Debug("can't fetch GAS balance", zap.String("key", keyHex), zap.Error(err))
		} else {
			exportBalancesGAS[keyHex] = balanceGAS
		}

		record, err := locodedb.Get(node.Locode)
		if err != nil {
			m.logger.Debug("can't fetch geoposition of node from the NeoFS network map",
				zap.String("key", keyHex),
				zap.String("locode", node.Locode),
				zap.Error(err),
			)
		} else {
			nodeLoc := nodeLocation{
				name: record.Location,
				long: strconv.FormatFloat(float64(record.Point.Longitude), 'f', 4, 32),
				lat:  strconv.FormatFloat(float64(record.Point.Latitude), 'f', 4, 32),
			}

			exportCountries[nodeLoc]++
		}

		balanceNotary, err := m.notaryBalanceFetcher.FetchNotary(scriptHash)
		if err != nil {
			m.logger.Debug("can't fetch notary balance of node from the NeoFS network map",
				zap.String("key", keyHex),
				zap.Error(err),
			)
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
		storageNodeGASBalances.WithLabelValues(k).Set(v)
	}

	storageNodeNotaryBalances.Reset()
	for k, v := range exportBalancesNotary {
		storageNodeNotaryBalances.WithLabelValues(k).Set(v)
	}
}

func (m *SideJob) logNodes(msg string, nodes []*Node) {
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

func (m *SideJob) processInnerRing(ir keys.PublicKeys) {
	exportBalances := make(map[string]float64, len(ir))

	for _, key := range ir {
		keyHex := hex.EncodeToString(key.Bytes())

		balance, err := m.balanceFetcher.Fetch(gas.Hash, key.GetScriptHash())
		if err != nil {
			m.logger.Debug("can't fetch GAS balance of the NeoFS Inner Ring member",
				zap.String("key", keyHex),
				zap.Error(err),
			)
			continue
		}

		exportBalances[keyHex] = balance
	}

	innerRingBalances.Reset()
	for k, v := range exportBalances {
		innerRingBalances.WithLabelValues(k).Set(v)
	}
}

func (m *SideJob) processProxyContract() {
	balance, err := m.balanceFetcher.Fetch(gas.Hash, *m.proxy)
	if err != nil {
		m.logger.Debug("can't fetch proxy contract balance", zap.Stringer("address", m.proxy), zap.Error(err))
		return
	}

	proxyBalance.Set(balance)
}

func (m *SideJob) processSideAlphabet(alphabet keys.PublicKeys) {
	exportNotaryBalances := make(map[string]float64, len(alphabet))

	for _, key := range alphabet {
		keyHex := hex.EncodeToString(key.Bytes())

		balanceNotary, err := m.notaryBalanceFetcher.FetchNotary(key.GetScriptHash())
		if err != nil {
			m.logger.Debug("can't fetch notary balance of the NeoFS Alphabet member", zap.String("key", keyHex), zap.Error(err))
		} else {
			exportNotaryBalances[keyHex] = balanceNotary
		}
	}

	alphabetNotaryBalances.Reset()
	for k, v := range exportNotaryBalances {
		alphabetNotaryBalances.WithLabelValues(k).Set(v)
	}
}

func (m *SideJob) processSideChainSupply() {
	balance, err := m.balanceFetcher.FetchTotalSupply(m.balance)
	if err != nil {
		m.logger.Debug("can't fetch balance contract total supply", zap.Stringer("address", m.balance), zap.Error(err))
		return
	}

	sideChainSupply.Set(balance)
}

func (m *SideJob) processContainersNumber() {
	total, err := m.cnrFetcher.Total()
	if err != nil {
		m.logger.Warn("can't fetch number of available containers", zap.Error(err))
		return
	}

	containersNumber.Set(float64(total))
}

func (m *SideJob) processChainHeight() uint32 {
	var minHeight uint32
	heightData := m.heightFetcher.FetchHeight()

	for _, d := range heightData {
		chainHeight.WithLabelValues(d.Host).Set(float64(d.Value))

		if minHeight == 0 || d.Value < minHeight {
			minHeight = d.Value
		}
	}

	return minHeight
}

func (m *SideJob) processChainState(height uint32) {
	if height == 0 {
		return
	}

	stateData := m.stateFetcher.FetchState(height)
	chainState.Reset()

	h := float64(height)

	for _, d := range stateData {
		chainState.WithLabelValues(d.Host, d.Value).Set(h)
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
