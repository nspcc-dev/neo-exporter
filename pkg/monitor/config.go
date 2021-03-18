package monitor

import (
	"time"

	"github.com/spf13/viper"
)

const (
	// contracts scripthash
	cfgNetmapContract = "contracts.netmap"
	cfgProxyContract  = "contracts.proxy"

	// private key to communicate with blockchain
	cfgKey = "key"

	// neo rpc node related config values
	cfgNeoRPCEndpoint    = "rpc.endpoint"
	cfgNeoRPCDialTimeout = "rpc.dial_timeout"

	// monitor prometheus expose config values
	cfgMetricsEndpoint = "metrics.endpoint"
	cfgMetricsInterval = "metrics.interval"

	// geoip related config values
	cfgGeoIPDialTimeout = "geoip.dial_timeout"
	cfgGeoIPEndpoint    = "geoip.endpoint"
	cfgGeoIPAccessKey   = "geoip.access_key"
)

func DefaultConfiguration(cfg *viper.Viper) {
	cfg.SetDefault(cfgNetmapContract, "")
	cfg.SetDefault(cfgProxyContract, "")

	cfg.SetDefault(cfgKey, "")

	cfg.SetDefault(cfgNeoRPCEndpoint, "")
	cfg.SetDefault(cfgNeoRPCDialTimeout, 5*time.Second)

	cfg.SetDefault(cfgMetricsEndpoint, ":16512")
	cfg.SetDefault(cfgMetricsInterval, 15*time.Minute)

	cfg.SetDefault(cfgGeoIPDialTimeout, 5*time.Second)
	cfg.SetDefault(cfgGeoIPEndpoint, "http://api.ipstack.com")
	cfg.SetDefault(cfgGeoIPAccessKey, "")
}
