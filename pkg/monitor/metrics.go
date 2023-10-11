package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	location  = "location"
	longitude = "longitude"
	latitude  = "latitude"
)

var (
	locationPresent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "neofs_net_monitor",
			Name:      "netmap",
			Help:      "Locations where NeoFS storage nodes are located",
		},
		[]string{
			location,
			longitude,
			latitude,
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

	alphabetGASBalances = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "neofs_net_monitor",
			Name:      "alphabet_balance",
			Help:      "Main chain GAS amount of alphabet nodes",
		},
		[]string{
			"key",
		},
	)

	alphabetNotaryBalances = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "neofs_net_monitor",
			Name:      "alphabet_balance_notary",
			Help:      "Side chain notary balance of alphabet nodes",
		},
		[]string{
			"key",
		},
	)

	storageNodeGASBalances = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "neofs_net_monitor",
			Name:      "sn_balance",
			Help:      "Side chain GAS amount of storage nodes",
		},
		[]string{
			"key",
		},
	)

	storageNodeNotaryBalances = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "neofs_net_monitor",
			Name:      "sn_balance_notary",
			Help:      "Side chain notary balance of storage nodes",
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

	mainChainSupply = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: "neofs_net_monitor",
			Name:      "main_chain_supply",
			Help:      "Main chain GAS amount of neofs contract",
		},
	)

	sideChainSupply = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: "neofs_net_monitor",
			Name:      "side_chain_supply",
			Help:      "Side chain total supply of balance contract",
		},
	)

	alphabetDivergence = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "neofs_net_monitor",
			Name:      "alphabet_divergence_count",
			Help:      "Number of unique alphabet keys in main chain and side chain",
		},
		[]string{
			"chain",
		},
	)

	alphabetMainDivergence = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "neofs_net_monitor",
			Name:      "alphabet_main_divergence",
			Help:      "Alphabet keys divergence in main chain",
		},
		[]string{
			"key",
		},
	)

	alphabetSideDivergence = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "neofs_net_monitor",
			Name:      "alphabet_side_divergence",
			Help:      "Alphabet keys divergence in side chain",
		},
		[]string{
			"key",
		},
	)

	containersNumber = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: "neofs_net_monitor",
			Name:      "containers_number",
			Help:      "Number of available containers",
		},
	)

	chainHeight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "neofs_net_monitor",
			Name:      "chain_height",
			Help:      "Chain height in blocks",
		},
		[]string{
			"host",
		},
	)

	chainState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "neofs_net_monitor",
			Name:      "chain_state",
			Help:      "Chain state hash in specific height",
		},
		[]string{
			"host", "hash",
		},
	)
)
