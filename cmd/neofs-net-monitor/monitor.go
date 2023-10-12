package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/rpcclient/unwrap"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm/stackitem"
	"github.com/nspcc-dev/neofs-contract/nns"
	rpcnns "github.com/nspcc-dev/neofs-contract/rpc/nns"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/locode"
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

	containerContract, err := getScriptHash(cfg, sideNeogoClient, "container.neofs", cfgContainerContract)
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

	balance, err = getScriptHash(cfg, sideNeogoClient, "balance.neofs", cfgBalanceContract)
	if err != nil {
		return nil, fmt.Errorf("balance contract is not available: %w", err)
	}

	proxyContract, err := getScriptHash(cfg, sideNeogoClient, "proxy.neofs", cfgProxyContract)
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

	geoFetcher := locode.New(
		locode.Prm{
			Path: cfg.GetString(cfgLocodeDB),
		},
	)

	return monitor.New(monitor.Args{
		Balance:        balance,
		Proxy:          proxy,
		Neofs:          neofs,
		Logger:         logger,
		Sleep:          cfg.GetDuration(cfgMetricsInterval),
		MetricsAddress: cfg.GetString(cfgMetricsEndpoint),
		GeoFetcher:     geoFetcher,
		AlpFetcher:     alphabetFetcher,
		NmFetcher:      nmFetcher,
		IRFetcher:      nmFetcher,
		SideBlFetcher:  sideBalanceFetcher,
		MainBlFetcher:  mainBalanceFetcher,
		CnrFetcher:     cnrFetcher,
	}), nil
}

const nnsContractID = 1

func getScriptHash(cfg *viper.Viper, cli *pool.Pool, nnsKey, configKey string) (sh util.Uint160, err error) {
	cs, err := cli.GetContractStateByID(nnsContractID)
	if err != nil {
		return sh, fmt.Errorf("NNS contract state: %w", err)
	}

	hash := cfg.GetString(configKey)
	if len(hash) == 0 {
		sh, err = nnsResolve(cli, cs.Hash, nnsKey)
		if err != nil {
			return sh, fmt.Errorf("NNS.resolve: %w", err)
		}
	} else {
		sh, err = util.Uint160DecodeStringLE(hash)
		if err != nil {
			return sh, fmt.Errorf("NNS u160 decode: %w", err)
		}
	}

	return sh, nil
}

func nnsResolve(c *pool.Pool, nnsHash util.Uint160, domain string) (util.Uint160, error) {
	item, err := unwrap.Item(c.Call(nnsHash, "resolve", domain, int64(nns.TXT)))
	if err != nil {
		return util.Uint160{}, fmt.Errorf("contract invocation: %w", err)
	}

	if _, ok := item.(stackitem.Null); ok {
		return util.Uint160{}, errors.New("NNS record is missing")
	}

	// Parse the result of resolving NNS record.
	// It works with multiple formats (corresponding to multiple NNS versions).
	// If array of hashes is provided, it returns only the first one.
	if arr, ok := item.Value().([]stackitem.Item); ok {
		if len(arr) == 0 {
			return util.Uint160{}, errors.New("NNS record is missing")
		}
		item = arr[0]
	}
	bs, err := item.TryBytes()
	if err != nil {
		return util.Uint160{}, fmt.Errorf("malformed response: %w", err)
	}
	return util.Uint160DecodeStringLE(string(bs))
}
