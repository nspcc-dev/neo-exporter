package main

import (
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	delimiter = "."

	// contracts scripthash.
	cfgNeoFSContract = "contracts.neofs"

	// neo rpc node related config values.
	prefix = "chain"

	cfgChainFSChain = "chain.fschain"

	cfgNeoRPCEndpoint        = "rpc.endpoint"
	cfgNeoRPCDialTimeout     = "rpc.dial_timeout"
	cfgNeoRPCRecheckInterval = "rpc.health_recheck_interval"

	// monitor prometheus expose config values.
	cfgMetricsEndpoint = "metrics.endpoint"
	cfgMetricsInterval = "metrics.interval"

	// level of logging.
	cfgLoggerLevel = "logger.level"
)

func DefaultConfiguration(cfg *viper.Viper) {
	cfg.SetDefault(prefix+delimiter+cfgNeoRPCEndpoint, "")
	cfg.SetDefault(prefix+delimiter+cfgNeoRPCDialTimeout, time.Minute)

	cfg.SetDefault(cfgMetricsEndpoint, ":16512")
	cfg.SetDefault(cfgMetricsInterval, 15*time.Second)

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
