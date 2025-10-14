package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	location  = "location"
	longitude = "longitude"
	latitude  = "latitude"
	namespace = "neo_exporter"
)

var (
	binaryVersion = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Help:      "Exporter version",
			Name:      "version",
			Namespace: namespace,
		},
		[]string{"version"},
	)

	locationPresent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
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
			Namespace: namespace,
			Name:      "netmap_dropped",
			Help:      "Amount of nodes that will be dropped from network in the next epoch",
		},
	)

	newNodesCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "netmap_new",
			Help:      "Amount of nodes that will be added to network in the next epoch",
		},
	)

	epochNumber = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "epoch",
			Help:      "Epoch number of NeoFS network",
		},
	)

	innerRingBalances = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "ir_balance",
			Help:      "Side chain GAS amount of inner ring nodes",
		},
		[]string{
			"key",
		},
	)

	alphabetGASBalances = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "alphabet_balance",
			Help:      "Main chain GAS amount of alphabet nodes",
		},
		[]string{
			"key",
		},
	)

	alphabetNotaryBalances = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "alphabet_balance_notary",
			Help:      "Side chain notary balance of alphabet nodes",
		},
		[]string{
			"key",
		},
	)

	storageNodeGASBalances = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "sn_balance",
			Help:      "Side chain GAS amount of storage nodes",
		},
		[]string{
			"key",
		},
	)

	storageNodeNotaryBalances = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "sn_balance_notary",
			Help:      "Side chain notary balance of storage nodes",
		},
		[]string{
			"key",
		},
	)

	proxyBalance = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "proxy_balance",
			Help:      "Side chain GAS amount of proxy contract",
		},
	)

	mainChainSupply = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "main_chain_supply",
			Help:      "Main chain GAS amount of neofs contract",
		},
	)

	fsChainSupply = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "fs_chain_supply",
			Help:      "FS chain total supply of balance contract",
		},
	)

	alphabetPubKeys = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "alphabet_public_key",
			Help:      "Alphabet public keys in chain",
		},
		[]string{
			"key",
		},
	)

	containersNumber = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "containers_number",
			Help:      "Number of available containers",
		},
	)

	chainHeight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "chain_height",
			Help:      "Chain height in blocks",
		},
		[]string{
			"host",
		},
	)

	chainState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "chain_state",
			Help:      "Chain state hash in specific height",
		},
		[]string{
			"host", "hash",
		},
	)

	nep17tracker = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "nep_17_balance",
			Help:      "NEP-17 balance of contract and account",
		},
		[]string{
			"symbol", "contract", "account",
		},
	)

	nep17trackerTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "nep_17_total_supply",
			Help:      "NEP-17 total supply of contract",
		},
		[]string{
			"symbol", "contract",
		},
	)

	candidateInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "candidate_info",
			Help:      "Candidate node info",
		},
		[]string{
			"host", "last_active_epoch",
		},
	)

	storageNodeCapacity = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "sn_capacity",
			Help:      "Storage node capacity (GB)",
		},
		[]string{
			"host", "key",
		},
	)

	storageNodeTotalCapacity = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "sn_capacity_total",
			Help:      "Storage nodes total capacity (GB)",
		},
	)
)

// RegisterFSChainMetrics inits prometheus metrics for side chain. Panics if can't do it.
func RegisterFSChainMetrics() {
	prometheus.MustRegister(binaryVersion)
	prometheus.MustRegister(locationPresent)
	prometheus.MustRegister(droppedNodesCount)
	prometheus.MustRegister(newNodesCount)
	prometheus.MustRegister(epochNumber)
	prometheus.MustRegister(storageNodeGASBalances)
	prometheus.MustRegister(storageNodeNotaryBalances)
	prometheus.MustRegister(innerRingBalances)
	prometheus.MustRegister(alphabetNotaryBalances)
	prometheus.MustRegister(proxyBalance)
	prometheus.MustRegister(fsChainSupply)
	prometheus.MustRegister(alphabetPubKeys) // used for both monitors
	prometheus.MustRegister(containersNumber)
	prometheus.MustRegister(chainHeight)
	prometheus.MustRegister(chainState)
	prometheus.MustRegister(nep17tracker)      // used for both monitors
	prometheus.MustRegister(nep17trackerTotal) // used for both monitors
	prometheus.MustRegister(candidateInfo)     // used for both monitors
	prometheus.MustRegister(storageNodeCapacity)
	prometheus.MustRegister(storageNodeTotalCapacity)
}

// RegisterMainChainMetrics inits prometheus metrics for main chain. Panics if can't do it.
func RegisterMainChainMetrics() {
	prometheus.MustRegister(binaryVersion)
	prometheus.MustRegister(alphabetGASBalances)
	prometheus.MustRegister(mainChainSupply)
	prometheus.MustRegister(alphabetPubKeys)   // used for both monitors
	prometheus.MustRegister(nep17tracker)      // used for both monitors
	prometheus.MustRegister(nep17trackerTotal) // used for both monitors
}

// SetExporterVersion sets neo-exporter version metric.
func SetExporterVersion(ver string) {
	binaryVersion.WithLabelValues(ver).Add(1)
}
