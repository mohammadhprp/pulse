package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mohammadhptp/pulse/internal/agent"
	"github.com/mohammadhptp/pulse/pkg/logger"
	"github.com/segmentio/kafka-go"
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

	logger.Info("Agent starting", zap.String("logLevel", logLevel))

	broker := viper.GetString("KAFKA_BROKER")
	topic := viper.GetString("KAFKA_TOPIC")

	// Validate required configuration
	if broker == "" {
		logger.Fatal("KAFKA_BROKER is not set")
	}
	if topic == "" {
		logger.Fatal("KAFKA_TOPIC is not set")
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{broker},
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
	})
	defer writer.Close()

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		sig := <-signals
		logger.Info("Shutdown signal received", zap.String("signal", sig.String()))
		cancel()
	}()

	logger.Info("Agent started", zap.String("broker", broker), zap.String("topic", topic))
	if err := agent.ProduceLogs(ctx, writer, os.Stdin); err != nil {
		logger.Fatal("Produce error", zap.Error(err))
	}
}
