package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/nspcc-dev/neofs-net-monitor/pkg/monitor"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const envPrefix = "neofs_net_monitor"

// Version is an application version.
var Version = "dev"

func main() {
	configFile := flag.String("config", "", "path to config")
	versionFlag := flag.Bool("version", false, "application version")
	flag.Parse()

	if *versionFlag {
		fmt.Println("version:", Version)
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	cfg, err := newConfig(*configFile)
	if err != nil {
		log.Printf("can't initialize application config: %s", err.Error())
		os.Exit(1)
	}

	neofsMonitor, err := monitor.New(ctx, cfg)
	if err != nil {
		log.Printf("can't initialize netmap monitor: %s", err.Error())
		os.Exit(1)
	}

	neofsMonitor.Start(ctx)

	neofsMonitor.Logger().Info("application started", zap.String("version", Version))

	<-ctx.Done()

	neofsMonitor.Stop()

	neofsMonitor.Logger().Info("application stopped", zap.String("version", Version))
}

func newConfig(path string) (*viper.Viper, error) {
	var (
		err error
		v   = viper.New()
	)

	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	monitor.DefaultConfiguration(v)

	if path != "" {
		v.SetConfigFile(path)
		v.SetConfigType("yml")
		err = v.ReadInConfig()
	}

	return v, err
}
