package main

import (
	"context"
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/util"
	rpcnns "github.com/nspcc-dev/neofs-contract/rpc/nns"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/monitor"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/morphchain"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/morphchain/contracts"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/multinodepool"
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

	sideChainEndpoints := cfg.GetStringSlice(sidePrefix + delimiter + cfgNeoRPCEndpoint)
	sideChainTimeout := cfg.GetDuration(sidePrefix + delimiter + cfgNeoRPCDialTimeout)
	sideChainRecheck := cfg.GetDuration(sidePrefix + delimiter + cfgNeoRPCRecheckInterval)

	mainChainEndpoints := cfg.GetStringSlice(mainPrefix + delimiter + cfgNeoRPCEndpoint)
	mainChainTimeout := cfg.GetDuration(mainPrefix + delimiter + cfgNeoRPCDialTimeout)
	mainChainRecheck := cfg.GetDuration(mainPrefix + delimiter + cfgNeoRPCRecheckInterval)

	sideNeogoClient, err := pool.NewPool(ctx, pool.PrmPool{
		Endpoints:       sideChainEndpoints,
		DialTimeout:     sideChainTimeout,
		RecheckInterval: sideChainRecheck,
	})
	if err != nil {
		return nil, fmt.Errorf("can't create side chain neo-go client: %w", err)
	}

	mainNeogoClient, err := pool.NewPool(ctx, pool.PrmPool{
		Endpoints:       mainChainEndpoints,
		DialTimeout:     mainChainTimeout,
		RecheckInterval: mainChainRecheck,
	})
	if err != nil {
		return nil, fmt.Errorf("can't create main chain neo-go client: %w", err)
	}

	netmapContract, err := sideNeogoClient.ResolveContract(rpcnns.NameNetmap)
	if err != nil {
		return nil, fmt.Errorf("can't read netmap scripthash: %w", err)
	}

	containerContract, err := sideNeogoClient.ResolveContract(rpcnns.NameContainer)
	if err != nil {
		return nil, fmt.Errorf("can't read container scripthash: %w", err)
	}

	nmFetcher, err := contracts.NewNetmap(contracts.NetmapArgs{
		Pool:           sideNeogoClient,
		NetmapContract: netmapContract,
		Logger:         logger,
	})
	if err != nil {
		return nil, fmt.Errorf("can't initialize netmap fetcher: %w", err)
	}

	cnrFetcher, err := contracts.NewContainer(sideNeogoClient, containerContract)
	if err != nil {
		return nil, fmt.Errorf("can't initialize container fetcher: %w", err)
	}

	alphabetFetcher, err := morphchain.NewAlphabetFetcher(morphchain.AlphabetFetcherArgs{
		Committeer: sideNeogoClient,
		Designater: mainNeogoClient,
	})
	if err != nil {
		return nil, fmt.Errorf("can't initialize alphabet fetcher: %w", err)
	}

	sideBalanceFetcher, err := morphchain.NewBalanceFetcher(morphchain.BalanceFetcherArgs{
		Cli: sideNeogoClient,
	})
	if err != nil {
		return nil, fmt.Errorf("can't initialize side balance fetcher: %w", err)
	}

	mainBalanceFetcher, err := morphchain.NewBalanceFetcher(morphchain.BalanceFetcherArgs{
		Cli: mainNeogoClient,
	})
	if err != nil {
		return nil, fmt.Errorf("can't initialize main balance fetcher: %w", err)
	}

	var (
		balance util.Uint160
		proxy   *util.Uint160
		neofs   *util.Uint160
	)

	balance, err = sideNeogoClient.ResolveContract(rpcnns.NameBalance)
	if err != nil {
		return nil, fmt.Errorf("balance contract is not available: %w", err)
	}

	proxyContract, err := sideNeogoClient.ResolveContract(rpcnns.NameProxy)
	if err != nil {
		logger.Info("proxy disabled")
	} else {
		proxy = &proxyContract
	}

	neofsContract := cfg.GetString(cfgNeoFSContract)
	if len(neofsContract) != 0 {
		sh, err := util.Uint160DecodeStringLE(neofsContract)
		if err != nil {
			return nil, fmt.Errorf("NNS u160 decode: %w", err)
		}
		neofs = &sh
	} else {
		logger.Info("neofs contract ignored")
	}

	mnPool := multinodepool.NewPool(sideChainEndpoints, cfg.GetDuration(cfgMetricsInterval))
	if err = mnPool.Dial(ctx); err != nil {
		return nil, fmt.Errorf("multinodepool: %w", err)
	}

	return monitor.New(monitor.Args{
		Balance:        balance,
		Proxy:          proxy,
		Neofs:          neofs,
		Logger:         logger,
		Sleep:          cfg.GetDuration(cfgMetricsInterval),
		MetricsAddress: cfg.GetString(cfgMetricsEndpoint),
		AlpFetcher:     alphabetFetcher,
		NmFetcher:      nmFetcher,
		IRFetcher:      nmFetcher,
		SideBlFetcher:  sideBalanceFetcher,
		MainBlFetcher:  mainBalanceFetcher,
		CnrFetcher:     cnrFetcher,
		HeightFetcher:  mnPool,
		StateFetcher:   mnPool,
	}), nil
}
