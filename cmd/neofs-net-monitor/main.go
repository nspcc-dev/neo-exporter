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

	ctx := gracefulContext()

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

	log.Println("application started")

	<-ctx.Done()

	neofsMonitor.Stop()

	log.Println("application stopped")
}

func gracefulContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		sig := <-ch
		log.Printf("ctx: received signal %s, closing", sig.String())
		cancel()
	}()

	return ctx
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
