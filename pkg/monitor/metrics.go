package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	countriesPresent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "neofs_net_monitor",
			Name:      "netmap",
			Help:      "Countries where NeoFS storage nodes are located",
		},
		[]string{
			"country",
		},
	)

	droppedNodesCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: "neofs_net_monitor",
			Name:      "netmap_dropped",
			Help:      "Amount of nodes that will be dropped from network in the next epoch",
		},
	)

	newNodesCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: "neofs_net_monitor",
			Name:      "netmap_new",
			Help:      "Amount of nodes that will be added to network in the next epoch",
		},
	)

	epochNumber = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: "neofs_net_monitor",
			Name:      "epoch",
			Help:      "Epoch number of NeoFS network",
		},
	)

	innerRingBalances = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "neofs_net_monitor",
			Name:      "ir_balance",
			Help:      "Side chain GAS amount of inner ring nodes",
		},
		[]string{
			"key",
		},
	)

	alphabetBalances = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "neofs_net_monitor",
			Name:      "alphabet_balance",
			Help:      "Main chain GAS amount of alphabet nodes",
		},
		[]string{
			"key",
		},
	)

	storageNodeBalances = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "neofs_net_monitor",
			Name:      "sn_balance",
			Help:      "Side chain GAS amount of storage nodes",
		},
		[]string{
			"key",
		},
	)

	proxyBalance = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: "neofs_net_monitor",
			Name:      "proxy_balance",
			Help:      "Side chain GAS amount of proxy contract",
		},
	)
)
