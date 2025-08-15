package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/NorskHelsenett/netbird-log-forwarder/cmd/settings"
	"github.com/NorskHelsenett/netbird-log-forwarder/internal/cache/netbird"
	"github.com/NorskHelsenett/netbird-log-forwarder/internal/logger"
	"github.com/NorskHelsenett/netbird-log-forwarder/internal/webserver"
	"github.com/spf13/viper"
)

func main() {

	// read config.json file
	err := settings.InitConfig()

	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	if err := logger.InitLogger("./logs"); err != nil {
		log.Fatalf("logger init failed: %v", err)
	}
	logger.Log.Infoln("zap logger initialized successfully")

	err = netbird.NewUserCache(viper.GetString("netbird_token"))

	if err != nil {
		logger.Log.Errorf("Failed to initialize user cache: %v\n", err)
		os.Exit(1)
	}

	err = netbird.NewPeerCache(viper.GetString("netbird_token"))
	if err != nil {
		logger.Log.Errorf("Failed to initialize peer cache: %v\n", err)
		os.Exit(1)
	}

	// fmt.Println("UserCache", userCache)
	// Example usage of userCache

	// fmt.Println("User cache initialized successfully")

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
	fmt.Printf("Received signal: %s. Shutting down...", sig)
	logger.Log.Infof("Received signal: %s. Shutting down...", sig)
	serverCancel()

}
