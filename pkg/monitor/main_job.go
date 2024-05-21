package monitor

import (
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/gas"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"go.uber.org/zap"
)

type (
	MainJobArgs struct {
		AlphabetFetcher AlphabetFetcher
		BalanceFetcher  Nep17BalanceFetcher
		Neofs           *util.Uint160
		Logger          *zap.Logger
		Nep17tracker    *Nep17tracker
	}

	MainJob struct {
		alphabetFetcher AlphabetFetcher
		balanceFetcher  Nep17BalanceFetcher
		logger          *zap.Logger
		neofs           *util.Uint160
		nep17tracker    *Nep17tracker
	}
)

func NewMainJob(args MainJobArgs) *MainJob {
	return &MainJob{
		alphabetFetcher: args.AlphabetFetcher,
		balanceFetcher:  args.BalanceFetcher,
		logger:          args.Logger,
		neofs:           args.Neofs,
		nep17tracker:    args.Nep17tracker,
	}
}

func (m *MainJob) Process() {
	if mainAlphabet, err := m.alphabetFetcher.FetchAlphabet(); err != nil {
		m.logger.Warn("can't read NeoFS Aphabet members", zap.Error(err))
	} else {
		processAlphabetPublicKeys(mainAlphabet)
		m.processMainAlphabet(mainAlphabet)
	}

	m.processMainChainSupply()
	m.processNep17tracker()
}

func (m *MainJob) processNep17tracker() {
	if m.nep17tracker != nil {
		m.nep17tracker.Process(nep17tracker, nep17trackerTotal)
	}
}

func (m *MainJob) processMainAlphabet(alphabet keys.PublicKeys) {
	exportGasBalances := make(map[string]float64, len(alphabet))

	for _, key := range alphabet {
		keyHex := key.StringCompressed()

		balanceGAS, err := m.balanceFetcher.Fetch(gas.Hash, key.GetScriptHash())
		if err != nil {
			m.logger.Debug("can't fetch gas balance", zap.String("key", keyHex), zap.Error(err))
		} else {
			exportGasBalances[keyHex] = balanceGAS
		}
	}

	alphabetGASBalances.Reset()
	for k, v := range exportGasBalances {
		alphabetGASBalances.WithLabelValues(k).Set(v)
	}
}

func (m *MainJob) processMainChainSupply() {
	if m.neofs == nil {
		return
	}

	balance, err := m.balanceFetcher.Fetch(gas.Hash, *m.neofs)
	if err != nil {
		m.logger.Debug("can't fetch NeoFS contract's GAS balance", zap.Error(err))
		return
	}

	mainChainSupply.Set(balance)
}
