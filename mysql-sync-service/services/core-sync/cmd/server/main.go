package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"mysql-sync-service/internal/api"
	"mysql-sync-service/internal/config"
	"mysql-sync-service/internal/logger"
	"mysql-sync-service/internal/store"
	"mysql-sync-service/internal/sync"
)

func main() {
	// Load Config
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Init Logger
	if err := logger.InitLogger(cfg.Logging.Level, cfg.Logging.Format); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Log.Info("Starting MySQL Sync Service")

	// Init State Store
	// For now, assume MySQL store
	stateStore, err := store.NewMySQLStore(cfg.StateStorage)
	if err != nil {
		logger.Log.Fatal("Failed to init state store", zap.Error(err))
	}
	defer stateStore.Close()

	// Init Sync Manager
	syncManager, err := sync.NewManager(cfg, stateStore)
	if err != nil {
		logger.Log.Fatal("Failed to init sync manager", zap.Error(err))
	}
	defer syncManager.Close()

	// Init API
	handler := api.NewHandler(syncManager)
	router := handler.Routes()

	// Start Server
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	go func() {
		logger.Log.Info("Server listening", zap.String("addr", serverAddr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Server failed", zap.Error(err))
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down server...")
	syncManager.Stop()
	// server.Shutdown(ctx) could be added here
}
