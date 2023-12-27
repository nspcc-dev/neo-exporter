package main

import (
	"context"
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/util"
	rpcnns "github.com/nspcc-dev/neofs-contract/rpc/nns"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/monitor"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/morphchain"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/morphchain/contracts"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/pool"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func New(ctx context.Context, cfg *viper.Viper) (*monitor.Monitor, error) {
	logConf := zap.NewProductionConfig()
	logConf.Level = WithLevel(cfg.GetString(cfgLoggerLevel))
	logger, err := logConf.Build()
	if err != nil {
		return nil, err
	}

	sideChainEndpoints := cfg.GetStringSlice(prefix + delimiter + cfgNeoRPCEndpoint)
	sideChainTimeout := cfg.GetDuration(prefix + delimiter + cfgNeoRPCDialTimeout)
	sideChainRecheck := cfg.GetDuration(prefix + delimiter + cfgNeoRPCRecheckInterval)

	sideNeogoClient, err := pool.NewPool(ctx, pool.PrmPool{
		Endpoints:       sideChainEndpoints,
		DialTimeout:     sideChainTimeout,
		RecheckInterval: sideChainRecheck,
	})
	if err != nil {
		return nil, fmt.Errorf("can't create side chain neo-go client: %w", err)
	}

	var job monitor.Job
	if cfg.GetBool(cfgChainFSChain) {
		monitor.RegisterSideChainMetrics()
		job, err = sideChainJob(sideNeogoClient, logger)
	} else {
		monitor.RegisterMainChainMetrics()
		job, err = mainChainJob(cfg, sideNeogoClient, logger)
	}

	if err != nil {
		return nil, err
	}

	return monitor.New(
		job,
		cfg.GetString(cfgMetricsEndpoint),
		cfg.GetDuration(cfgMetricsInterval),
		logger,
	), nil
}

func mainChainJob(cfg *viper.Viper, neogoClient *pool.Pool, logger *zap.Logger) (*monitor.MainJob, error) {
	alphabetFetcher := morphchain.NewMainChainAlphabetFetcher(neogoClient)

	balanceFetcher, err := morphchain.NewBalanceFetcher(
		morphchain.BalanceFetcherArgs{
			Cli: neogoClient,
		})
	if err != nil {
		return nil, fmt.Errorf("can't initialize Neo chain balance reader: %w", err)
	}

	var neofs *util.Uint160

	neofsContract := cfg.GetString(cfgNeoFSContract)
	if len(neofsContract) != 0 {
		sh, err := util.Uint160DecodeStringLE(neofsContract)
		if err != nil {
			return nil, fmt.Errorf("decode configured NeoFS contract address %q: %w", cfgNeoFSContract, err)
		}
		neofs = &sh
	} else {
		logger.Info("NeoFS contract address not configured, continue without it")
	}

	return monitor.NewMainJob(monitor.MainJobArgs{
		AlphabetFetcher: alphabetFetcher,
		BalanceFetcher:  balanceFetcher,
		Neofs:           neofs,
		Logger:          logger,
	}), nil
}

func sideChainJob(neogoClient *pool.Pool, logger *zap.Logger) (*monitor.SideJob, error) {
	netmapContract, err := neogoClient.ResolveContract(rpcnns.NameNetmap)
	if err != nil {
		return nil, fmt.Errorf("can't read netmap scripthash: %w", err)
	}

	containerContract, err := neogoClient.ResolveContract(rpcnns.NameContainer)
	if err != nil {
		return nil, fmt.Errorf("can't read container scripthash: %w", err)
	}

	nmFetcher, err := contracts.NewNetmap(contracts.NetmapArgs{
		Pool:           neogoClient,
		NetmapContract: netmapContract,
		Logger:         logger,
	})
	if err != nil {
		return nil, fmt.Errorf("can't initialize netmap fetcher: %w", err)
	}

	cnrFetcher, err := contracts.NewContainer(neogoClient, containerContract)
	if err != nil {
		return nil, fmt.Errorf("can't initialize container fetcher: %w", err)
	}

	alphabetFetcher := morphchain.NewSideChainAlphabetFetcher(neogoClient)

	balanceFetcher, err := morphchain.NewBalanceFetcher(morphchain.BalanceFetcherArgs{
		Cli: neogoClient,
	})
	if err != nil {
		return nil, fmt.Errorf("can't initialize side balance fetcher: %w", err)
	}

	var (
		balance util.Uint160
		proxy   *util.Uint160
	)

	balance, err = neogoClient.ResolveContract(rpcnns.NameBalance)
	if err != nil {
		return nil, fmt.Errorf("balance contract is not available: %w", err)
	}

	proxyContract, err := neogoClient.ResolveContract(rpcnns.NameProxy)
	if err != nil {
		logger.Info("proxy disabled")
	} else {
		proxy = &proxyContract
	}

	return monitor.NewSideJob(monitor.SideJobArgs{
		Logger:          logger,
		Balance:         balance,
		Proxy:           proxy,
		AlphabetFetcher: alphabetFetcher,
		NmFetcher:       nmFetcher,
		IRFetcher:       nmFetcher,
		BalanceFetcher:  balanceFetcher,
		CnrFetcher:      cnrFetcher,
		HeightFetcher:   neogoClient,
		StateFetcher:    neogoClient,
	}), nil
}
