package monitor

import (
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	delimiter = "."

	// contracts scripthash
	cfgNetmapContract    = "contracts.netmap"
	cfgProxyContract     = "contracts.proxy"
	cfgBalanceContract   = "contracts.balance"
	cfgNeoFSContract     = "contracts.neofs"
	cfgContainerContract = "contracts.container"

	// private key to communicate with blockchain
	cfgKey = "key"

	// neo rpc node related config values
	mainPrefix = "mainnet"
	sidePrefix = "morph"

	cfgNeoRPCEndpoint        = "rpc.endpoint"
	cfgNeoRPCDialTimeout     = "rpc.dial_timeout"
	cfgNeoRPCRecheckInterval = "rpc.health_recheck_interval"

	// monitor prometheus expose config values
	cfgMetricsEndpoint = "metrics.endpoint"
	cfgMetricsInterval = "metrics.interval"

	// path to the NeoFS locode database
	cfgLocodeDB = "locode.db.path"

	// level of logging
	cfgLoggerLevel = "logger.level"
)

func DefaultConfiguration(cfg *viper.Viper) {
	cfg.SetDefault(cfgNetmapContract, "")
	cfg.SetDefault(cfgProxyContract, "")

	cfg.SetDefault(cfgKey, "")

	cfg.SetDefault(sidePrefix+delimiter+cfgNeoRPCEndpoint, "")
	cfg.SetDefault(sidePrefix+delimiter+cfgNeoRPCDialTimeout, 5*time.Second)

	cfg.SetDefault(mainPrefix+delimiter+cfgNeoRPCEndpoint, "")
	cfg.SetDefault(mainPrefix+delimiter+cfgNeoRPCDialTimeout, 5*time.Second)

	cfg.SetDefault(cfgMetricsEndpoint, ":16512")
	cfg.SetDefault(cfgMetricsInterval, 15*time.Minute)

	cfg.SetDefault(cfgLocodeDB, "./locode/db")

	cfg.SetDefault(cfgLoggerLevel, "info")
}

func WithLevel(level string) zap.AtomicLevel {
	return safeLevel(level)
}

func safeLevel(lvl string) zap.AtomicLevel {
	switch strings.ToLower(lvl) {
	case "debug":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case "warn":
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "fatal":
		return zap.NewAtomicLevelAt(zap.FatalLevel)
	case "panic":
		return zap.NewAtomicLevelAt(zap.PanicLevel)
	default:
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	}
}
