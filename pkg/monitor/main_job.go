package monitor

import (
	"encoding/hex"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"go.uber.org/zap"
)

type (
	MainJobArgs struct {
		AlphabetFetcher AlphabetFetcher
		BalanceFetcher  BalanceFetcher
		Neofs           *util.Uint160
		Logger          *zap.Logger
	}

	MainJob struct {
		alphabetFetcher AlphabetFetcher
		balanceFetcher  BalanceFetcher
		logger          *zap.Logger
		neofs           *util.Uint160
	}
)

func NewMainJob(args MainJobArgs) *MainJob {
	return &MainJob{
		alphabetFetcher: args.AlphabetFetcher,
		balanceFetcher:  args.BalanceFetcher,
		logger:          args.Logger,
		neofs:           args.Neofs,
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
}

func (m *MainJob) processMainAlphabet(alphabet keys.PublicKeys) {
	exportGasBalances := make(map[string]int64, len(alphabet))

	for _, key := range alphabet {
		keyHex := hex.EncodeToString(key.Bytes())

		balanceGAS, err := m.balanceFetcher.FetchGAS(*key)
		if err != nil {
			m.logger.Debug("can't fetch gas balance", zap.String("key", keyHex), zap.Error(err))
		} else {
			exportGasBalances[keyHex] = balanceGAS
		}
	}

	alphabetGASBalances.Reset()
	for k, v := range exportGasBalances {
		alphabetGASBalances.WithLabelValues(k).Set(float64(v))
	}
}

func (m *MainJob) processMainChainSupply() {
	if m.neofs == nil {
		return
	}

	balance, err := m.balanceFetcher.FetchGASByScriptHash(*m.neofs)
	if err != nil {
		m.logger.Debug("can't fetch NeoFS contract's GAS balance", zap.Error(err))
		return
	}

	mainChainSupply.Set(float64(balance))
}
