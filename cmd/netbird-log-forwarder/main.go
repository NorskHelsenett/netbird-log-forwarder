package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/NorskHelsenett/netbird-log-forwarder/cmd/settings"
	"github.com/NorskHelsenett/netbird-log-forwarder/internal/cache/netbird"
	"github.com/NorskHelsenett/netbird-log-forwarder/internal/logger"
	"github.com/NorskHelsenett/netbird-log-forwarder/internal/webserver"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func main() {

	if err := logger.InitLogger("./logs"); err != nil {
		log.Fatalf("logger init failed: %v", err)
	}
	logger.Log.Infoln("Zap logger initialized successfully")

	used, err := settings.InitConfig("./config.yaml")
	if err != nil {
		logger.Log.Fatalf("config init failed: %v", err)
	}
	logger.Log.Infof("Config file %s loaded successfully", used)

	netbirdToken := viper.GetString("netbird.token")
	err = netbird.NewUserCache(netbirdToken)
	if err != nil {
		logger.Log.Errorf("Failed to initialize user cache: %v\n", err)
		os.Exit(1)
	}

	err = netbird.NewPeerCache(netbirdToken)
	if err != nil {
		logger.Log.Errorf("Failed to initialize peer cache: %v\n", err)
		os.Exit(1)
	}

	// Watch for config changes
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		logger.Log.Infof("Config reloaded: %s", e.Name)
		_, _ = settings.InitConfig("./config.yaml")
		if err != nil {
			log.Fatalf("Config reload failed: %v", err)
		}
	})

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start web server in a goroutine
	_, serverCancel := context.WithCancel(context.Background())
	go func() {
		webserver.InitHttpServer()
	}()

	// Wait for termination signal
	sig := <-sigChan
	logger.Log.Infof("Received signal: %s. Shutting down...", sig)
	serverCancel()

}
