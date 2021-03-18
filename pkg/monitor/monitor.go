package monitor

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/util"
	crypto "github.com/nspcc-dev/neofs-crypto"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/geoip"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/morphchain"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

type (
	GeoIPFetcher interface {
		Fetch(string) (geoip.Info, error)
	}

	NetmapFetcher interface {
		FetchNetmap() (morphchain.NetmapInfo, error)
	}

	InnerRingFetcher interface {
		FetchInnerRingKeys() (keys.PublicKeys, error)
	}

	BalanceFetcher interface {
		FetchGAS(keys.PublicKey) (int64, error)
		FetchGASByScriptHash(uint160 util.Uint160) (int64, error)
	}

	Monitor struct {
		proxy         util.Uint160
		sleep         time.Duration
		metricsServer http.Server
		ipFetcher     GeoIPFetcher
		nmFetcher     NetmapFetcher
		irFetcher     InnerRingFetcher
		blFetcher     BalanceFetcher
	}
)

func New(ctx context.Context, cfg *viper.Viper) (*Monitor, error) {
	key, err := readKey(cfg)
	if err != nil {
		return nil, err
	}

	ipFetcher, err := geoip.NewCachedFetcher(geoip.FetcherArgs{
		Timeout:   cfg.GetDuration(cfgGeoIPDialTimeout),
		Endpoint:  cfg.GetString(cfgGeoIPEndpoint),
		AccessKey: cfg.GetString(cfgGeoIPAccessKey),
	})
	if err != nil {
		return nil, fmt.Errorf("can't initialize geoip fetcher: %w", err)
	}

	netmapContract, err := util.Uint160DecodeStringLE(cfg.GetString(cfgNetmapContract))
	if err != nil {
		return nil, fmt.Errorf("can't read netmap scripthash: %w", err)
	}

	nmFetcher, err := morphchain.NewNetmapFetcher(ctx, morphchain.NetmapFetcherArgs{
		Key:            key,
		Endpoint:       cfg.GetString(cfgNeoRPCEndpoint),
		DialTimeout:    cfg.GetDuration(cfgNeoRPCDialTimeout),
		NetmapContract: netmapContract,
	})
	if err != nil {
		return nil, fmt.Errorf("can't initialize netmap fetcher: %w", err)
	}

	balanceFetcher, err := morphchain.NewBalanceFetcher(ctx, morphchain.BalanceFetcherArgs{
		Endpoint:    cfg.GetString(cfgNeoRPCEndpoint),
		DialTimeout: cfg.GetDuration(cfgNeoRPCDialTimeout),
	})
	if err != nil {
		return nil, fmt.Errorf("can't initialize balance fetcher: %w", err)
	}

	proxyContract, err := util.Uint160DecodeStringLE(cfg.GetString(cfgProxyContract))
	if err != nil {
		return nil, fmt.Errorf("can't read proxy scripthash: %w", err)
	}

	return &Monitor{
		proxy: proxyContract,
		sleep: cfg.GetDuration(cfgMetricsInterval),
		metricsServer: http.Server{
			Addr:    cfg.GetString(cfgMetricsEndpoint),
			Handler: promhttp.Handler(),
		},
		ipFetcher: ipFetcher,
		nmFetcher: nmFetcher,
		irFetcher: nmFetcher,
		blFetcher: balanceFetcher,
	}, nil
}

func (m *Monitor) Start(ctx context.Context) {
	prometheus.MustRegister(countriesPresent)
	prometheus.MustRegister(epochNumber)
	prometheus.MustRegister(storageNodeBalances)
	prometheus.MustRegister(innerRingBalances)
	prometheus.MustRegister(proxyBalance)

	go func() {
		err := m.metricsServer.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("monitor: strart metrics server error %s", err.Error())
		}
	}()

	go m.Job(ctx)
}

func (m *Monitor) Stop() {
	err := m.metricsServer.Close()
	if err != nil {
		log.Printf("monitor: stop metrics server error %s", err.Error())
	}
}

func (m *Monitor) Job(ctx context.Context) {
	for {
		log.Println("monitor: scraping data from side chain")

		netmap, err := m.nmFetcher.FetchNetmap()
		if err != nil {
			log.Printf("monitor: can't scrap network map info, %s", err)
		} else {
			m.processNetworkMap(netmap)
		}

		innerRing, err := m.irFetcher.FetchInnerRingKeys()
		if err != nil {
			log.Printf("monitor: can't scrap inner ring info, %s", err)
		} else {
			m.processInnerRing(innerRing)
		}

		m.processProxyContract()

		select {
		case <-time.After(m.sleep):
			// sleep for some time before next prometheus update
		case <-ctx.Done():
			log.Println("monitor: context closed, stop monitor")
			return
		}
	}
}

func (m *Monitor) processNetworkMap(nm morphchain.NetmapInfo) {
	exportCountries := make(map[string]int, len(nm.Addresses))
	exportBalances := make(map[string]int64, len(nm.Addresses))

	for _, addr := range nm.Addresses {
		info, err := m.ipFetcher.Fetch(addr)
		if err != nil {
			log.Printf("monitor: can't fetch %s info, %s", addr, err)
			continue
		}

		exportCountries[info.CountryCode]++
	}

	for _, key := range nm.PublicKeys {
		keyHex := hex.EncodeToString(key.Bytes())

		balance, err := m.blFetcher.FetchGAS(*key)
		if err != nil {
			log.Printf("monitor: can't fetch %s balance, %s", keyHex, err)
			continue
		}

		exportBalances[keyHex] = balance
	}

	epochNumber.Set(float64(nm.Epoch))

	countriesPresent.Reset()
	for k, v := range exportCountries {
		countriesPresent.WithLabelValues(k).Set(float64(v))
	}

	storageNodeBalances.Reset()
	for k, v := range exportBalances {
		storageNodeBalances.WithLabelValues(k).Set(float64(v))
	}
}

func (m *Monitor) processInnerRing(ir keys.PublicKeys) {
	exportBalances := make(map[string]int64, len(ir))

	for _, key := range ir {
		keyHex := hex.EncodeToString(key.Bytes())

		balance, err := m.blFetcher.FetchGAS(*key)
		if err != nil {
			log.Printf("monitor: can't fetch %s balance, %s", keyHex, err)
			continue
		}

		exportBalances[keyHex] = balance
	}

	innerRingBalances.Reset()
	for k, v := range exportBalances {
		innerRingBalances.WithLabelValues(k).Set(float64(v))
	}
}

func (m *Monitor) processProxyContract() {
	balance, err := m.blFetcher.FetchGASByScriptHash(m.proxy)
	if err != nil {
		log.Printf("monitor: can't fetch proxy contract balance, %s", err)
		return
	}

	proxyBalance.Set(float64(balance))
}

func readKey(cfg *viper.Viper) (*ecdsa.PrivateKey, error) {
	key, err := crypto.LoadPrivateKey(cfg.GetString(cfgKey))
	if err == nil {
		log.Println("monitor: using private key from the config")
		return key, nil
	}

	log.Println("monitor: using random private key")

	buf := make([]byte, crypto.PrivateKeyCompressedSize)
	_, err = rand.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("can't generate private key: %w", err)
	}

	return crypto.UnmarshalPrivateKey(buf)
}
