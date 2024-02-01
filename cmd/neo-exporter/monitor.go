package main

import (
	"context"
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/util"
	rpcnns "github.com/nspcc-dev/neofs-contract/rpc/nns"
	"github.com/nspcc-dev/neo-exporter/pkg/model"
	"github.com/nspcc-dev/neo-exporter/pkg/monitor"
	"github.com/nspcc-dev/neo-exporter/pkg/morphchain"
	"github.com/nspcc-dev/neo-exporter/pkg/morphchain/contracts"
	"github.com/nspcc-dev/neo-exporter/pkg/pool"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(ctx context.Context, cfg *viper.Viper) (*monitor.Monitor, error) {
	logConf := zap.NewProductionConfig()
	logConf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logConf.Level = WithLevel(cfg.GetString(cfgLoggerLevel))
	logger, err := logConf.Build()
	if err != nil {
		return nil, err
	}

	zap.ReplaceGlobals(logger)

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
		job, err = sideChainJob(cfg, sideNeogoClient, logger)
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

	balanceFetcher, err := monitor.NewNep17BalanceFetcher(neogoClient)
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

	var items []model.Nep17Balance
	if err = cfg.UnmarshalKey("nep17", &items); err != nil {
		return nil, fmt.Errorf("cfg nep17 parse: %w", err)
	}

	tasks, err := monitor.ParseNep17Tasks(balanceFetcher, items, &contracts.NNSNoOp{})
	if err != nil {
		return nil, err
	}

	nep17tracker, err := monitor.NewNep17tracker(balanceFetcher, tasks)
	if err != nil {
		return nil, fmt.Errorf("nep17tracker: %w", err)
	}

	return monitor.NewMainJob(monitor.MainJobArgs{
		AlphabetFetcher: alphabetFetcher,
		BalanceFetcher:  balanceFetcher,
		Neofs:           neofs,
		Logger:          logger,
		Nep17tracker:    nep17tracker,
	}), nil
}

func sideChainJob(cfg *viper.Viper, neogoClient *pool.Pool, logger *zap.Logger) (*monitor.SideJob, error) {
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

	balanceFetcher, err := monitor.NewNep17BalanceFetcher(neogoClient)
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

	var items []model.Nep17Balance
	if err = cfg.UnmarshalKey("nep17", &items); err != nil {
		return nil, fmt.Errorf("cfg nep17 parse: %w", err)
	}

	nnsHash, err := rpcnns.InferHash(neogoClient)
	if err != nil {
		return nil, fmt.Errorf("can't read nns scripthash: %w", err)
	}

	nnsContract, err := contracts.NewNNS(neogoClient, nnsHash)
	if err != nil {
		return nil, fmt.Errorf("can't initialize nns fetcher: %w", err)
	}

	tasks, err := monitor.ParseNep17Tasks(balanceFetcher, items, nnsContract)
	if err != nil {
		return nil, err
	}

	nep17tracker, err := monitor.NewNep17tracker(balanceFetcher, tasks)
	if err != nil {
		return nil, fmt.Errorf("nep17tracker: %w", err)
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
		Nep17tracker:    nep17tracker,
	}), nil
}
