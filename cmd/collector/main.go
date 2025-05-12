package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/mohammadhptp/pulse/internal/collector"
	"github.com/mohammadhptp/pulse/pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		logger.Warn("Config file not found or invalid", zap.Error(err))
		// Continue with environment variables
	}

	// Get log level from config, default to "info"
	logLevel := viper.GetString("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	// Initialize logger with configured level
	logger.InitLogger(logLevel)
	defer logger.Sync()

	logger.Info("Collector starting", zap.String("logLevel", logLevel))

	// Verify required configuration
	if viper.GetString("KAFKA_BROKER") == "" {
		logger.Fatal("KAFKA_BROKER is not set")
	}
	if viper.GetString("KAFKA_TOPIC") == "" {
		logger.Fatal("KAFKA_TOPIC is not set")
	}
	if viper.GetString("CLICKHOUSE_ADDR") == "" {
		logger.Fatal("CLICKHOUSE_ADDR is not set")
	}

	// Set up signal handling for graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("Collector started",
		zap.String("broker", viper.GetString("KAFKA_BROKER")),
		zap.String("topic", viper.GetString("KAFKA_TOPIC")),
		zap.String("clickhouse", viper.GetString("CLICKHOUSE_ADDR")))

	go collector.Run()

	// Wait for shutdown signal
	sig := <-signals
	logger.Info("Shutdown signal received", zap.String("signal", sig.String()))
}
