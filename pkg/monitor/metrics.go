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
)
