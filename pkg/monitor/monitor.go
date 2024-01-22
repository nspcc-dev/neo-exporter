package monitor

import (
	"context"
	"encoding/hex"
	"errors"
	"net/http"
	"sort"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type (
	Nep17BalanceFetcher interface {
		Fetch(tokenHash util.Uint160, account util.Uint160) (int64, error)
		FetchTotalSupply(tokenHash util.Uint160) (int64, error)
		Symbol(tokenHash util.Uint160) (string, error)
	}

	AlphabetFetcher interface {
		FetchAlphabet() (keys.PublicKeys, error)
	}

	Monitor struct {
		job           Job
		logger        *zap.Logger
		sleep         time.Duration
		metricsServer http.Server
	}

	Job interface {
		Process()
	}
)

func New(job Job, metricAddress string, sleep time.Duration, logger *zap.Logger) *Monitor {
	return &Monitor{
		job:    job,
		sleep:  sleep,
		logger: logger,
		metricsServer: http.Server{
			Addr:    metricAddress,
			Handler: promhttp.Handler(),
		},
	}
}

func (m *Monitor) Start(ctx context.Context) {
	go func() {
		err := m.metricsServer.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			m.logger.Error("start metrics server error", zap.Error(err))
		}
	}()

	go m.Job(ctx)
}

func (m *Monitor) Stop() {
	err := m.metricsServer.Close()
	if err != nil {
		m.logger.Error("stop metrics server error", zap.Error(err))
	}
}

func (m *Monitor) Job(ctx context.Context) {
	for {
		m.job.Process()

		select {
		case <-time.After(m.sleep):
			// sleep for some time before next prometheus update
		case <-ctx.Done():
			m.logger.Info("context closed, stop monitor")
			return
		}
	}
}

func (m *Monitor) Logger() *zap.Logger {
	return m.logger
}

func sortedAlphabet(alphabet keys.PublicKeys) []string {
	sort.Sort(alphabet)
	sorted := make([]string, 0, len(alphabet))
	for _, key := range alphabet {
		keyHex := hex.EncodeToString(key.Bytes())
		sorted = append(sorted, keyHex)
	}
	return sorted
}

func processAlphabetPublicKeys(alphabet keys.PublicKeys) {
	sorted := sortedAlphabet(alphabet)

	alphabetPubKeys.Reset()
	for _, key := range sorted {
		alphabetPubKeys.WithLabelValues(key).Set(1)
	}
}
