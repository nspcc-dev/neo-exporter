package monitor

import (
	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type (
	// Nep17tracker allows to get balances of accounts in corresponding contracts.
	Nep17tracker struct {
		balanceFetcher Nep17BalanceFetcher
		tasks          []Item
	}
)

// NewNep17tracker is a constructor for [Nep17tracker].
func NewNep17tracker(balanceFetcher Nep17BalanceFetcher, tasks []Item) (*Nep17tracker, error) {
	return &Nep17tracker{
		balanceFetcher: balanceFetcher,
		tasks:          tasks,
	}, nil
}

// Process runs the tasks and updates metrics.
func (n *Nep17tracker) Process(metric *prometheus.GaugeVec, metricTotal *prometheus.GaugeVec) {
	for _, item := range n.tasks {
		for _, acc := range item.Accounts {
			balance, err := n.balanceFetcher.Fetch(item.Hash, acc)
			if err != nil {
				zap.L().Error(
					"nep17 balance",
					zap.Error(err),
					zap.String("contract", item.Hash.StringLE()),
					zap.String("account", address.Uint160ToString(acc)),
				)
				continue
			}

			metric.WithLabelValues(
				item.Label,
				item.Symbol,
				item.Hash.StringLE(),
				address.Uint160ToString(acc),
			).Set(balance)
		}

		if item.Total {
			balance, err := n.balanceFetcher.FetchTotalSupply(item.Hash)
			if err != nil {
				zap.L().Error(
					"nep17 total balance",
					zap.Error(err),
					zap.String("contract", item.Hash.StringLE()),
				)
				continue
			}

			metricTotal.WithLabelValues(
				item.Label,
				item.Symbol,
				item.Hash.StringLE(),
			).Set(balance)
		}
	}
}
