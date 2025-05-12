package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mohammadhptp/pulse/internal/agent"
	"github.com/mohammadhptp/pulse/pkg/logger"
	"github.com/mohammadhptp/pulse/pkg/transport"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		logger.Warn("Config file not found or invalid", zap.Error(err))
	}

	logLevel := viper.GetString("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	logger.InitLogger(logLevel)
	defer logger.Sync()

	logger.Info("Agent starting", zap.String("logLevel", logLevel))

	broker := viper.GetString("KAFKA_BROKER")
	topic := viper.GetString("KAFKA_TOPIC")

	if broker == "" {
		logger.Fatal("KAFKA_BROKER is not set")
	}
	if topic == "" {
		logger.Fatal("KAFKA_TOPIC is not set")
	}

	httpPort := viper.GetInt("HTTP_PORT")
	if httpPort == 0 {
		logger.Fatal("HTTP_PORT is not set")
	}
	httpEndpoint := viper.GetString("HTTP_ENDPOINT")
	if httpEndpoint == "" {
		logger.Fatal("HTTP_ENDPOINT is not set")
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{broker},
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
	})
	defer writer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signals
		logger.Info("Shutdown signal received", zap.String("signal", sig.String()))
		cancel()
	}()

	httpTransport := transport.NewHTTPTransport(httpPort, httpEndpoint)

	processor := agent.NewEventProcessor(writer, httpTransport)

	logger.Info("Agent started",
		zap.String("broker", broker),
		zap.String("topic", topic),
		zap.Int("httpPort", httpPort),
		zap.String("httpEndpoint", httpEndpoint))

	if err := processor.Start(ctx); err != nil && err != context.Canceled {
		logger.Fatal("Event processor error", zap.Error(err))
	}
}
