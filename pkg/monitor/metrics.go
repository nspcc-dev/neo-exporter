package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	countriesPresent = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "neofs_net_monitor",
		Name:      "netmap",
		Help:      "Countries where NeoFS storage nodes are located",
	},
		[]string{
			"country",
		})

	epochNumber = prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: "neofs_net_monitor",
		Name:      "epoch",
		Help:      "Epoch number of NeoFS network",
	})

	innerRingBalances = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "neofs_net_monitor",
		Name:      "ir_balance",
		Help:      "Side chain GAS amount of inner ring nodes",
	},
		[]string{
			"key",
		})

	storageNodeBalances = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "neofs_net_monitor",
		Name:      "sn_balance",
		Help:      "Side chain GAS amount of storage nodes",
	},
		[]string{
			"key",
		})

	proxyBalance = prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: "neofs_net_monitor",
		Name:      "proxy_balance",
		Help:      "Side chain GAS amount of proxy contract",
	})
)
