package monitor

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

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

	Monitor struct {
		sleep         time.Duration
		metricsServer http.Server
		ipFetcher     GeoIPFetcher
		nmFetcher     NetmapFetcher
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

	contract, err := util.Uint160DecodeStringLE(cfg.GetString(cfgNetmapContract))
	if err != nil {
		return nil, fmt.Errorf("can't read netmap scripthash: %w", err)
	}

	nmFetcher, err := morphchain.NewNetmapFetcher(ctx, morphchain.NetmapFetcherArgs{
		Key:            key,
		Endpoint:       cfg.GetString(cfgNeoRPCEndpoint),
		DialTimeout:    cfg.GetDuration(cfgNeoRPCDialTimeout),
		NetmapContract: contract,
	})
	if err != nil {
		return nil, fmt.Errorf("can't initialize netmap fetcher: %w", err)
	}

	return &Monitor{
		sleep: cfg.GetDuration(cfgMetricsInterval),
		metricsServer: http.Server{
			Addr:    cfg.GetString(cfgMetricsEndpoint),
			Handler: promhttp.Handler(),
		},
		ipFetcher: ipFetcher,
		nmFetcher: nmFetcher,
	}, nil
}

func (m *Monitor) Start(ctx context.Context) {
	prometheus.MustRegister(countriesPresent)
	prometheus.MustRegister(epochNumber)

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
			log.Printf("monitor: can't scrap network map info, %s", err.Error())
		} else {
			m.processNetworkMap(netmap)
		}

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
	exportData := make(map[string]int, len(nm.Addresses))

	for _, addr := range nm.Addresses {
		info, err := m.ipFetcher.Fetch(addr)
		if err != nil {
			log.Printf("monitor: can't fetch %s info, %s", addr, err)
		}

		exportData[info.CountryCode]++
	}

	epochNumber.Set(float64(nm.Epoch))
	countriesPresent.Reset()
	for k, v := range exportData {
		countriesPresent.WithLabelValues(k).Set(float64(v))
	}
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
