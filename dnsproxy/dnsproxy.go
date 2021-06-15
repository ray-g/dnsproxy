package dnsproxy

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/ray-g/dnsproxy/api"
	"github.com/ray-g/dnsproxy/blocker"
	mem "github.com/ray-g/dnsproxy/cache/memcache"
	conf "github.com/ray-g/dnsproxy/config"
	"github.com/ray-g/dnsproxy/logger"
	r "github.com/ray-g/dnsproxy/resolver"
	"github.com/ray-g/dnsproxy/stats"
)

func Serve(filepath string) {
	config, err := conf.LoadConfig(filepath)
	if err != nil {
		logger.Fatal(err)
	}

	logger.InitLogger("DNSProxy", config.DebugMode)

	cache := mem.NewCache()
	dnshandler := r.NewHandler(&config.Resolver, cache)
	dnsserver := r.NewServer(config.DNSServer.BindAddr, dnshandler)
	dnsserver.Run()

	blocker.PerformUpdate(&config.Blocker, cache, false)

	if config.APIServer.Enable {
	err = api.StartAPIServer(config.APIServer.BindAddr, config.DebugMode, cache)
	if err != nil {
		logger.Fatalf("Cannot start the API server %s", err)
	}
	}

	stats.Activate()

	// Waiting for close
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGHUP)
	sig := <-osSignals
	logger.Debugf("Received signal: %v", sig)
}
