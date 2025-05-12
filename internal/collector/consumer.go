package collector

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/mohammadhptp/pulse/internal/storage"
	"github.com/mohammadhptp/pulse/pkg/logger"
	"github.com/mohammadhptp/pulse/pkg/models"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func Run() {
	broker := viper.GetString("KAFKA_BROKER")
	topic := viper.GetString("KAFKA_TOPIC")

	// Validate config
	if broker == "" {
		logger.Fatal("KAFKA_BROKER not set in configuration")
	}
	if topic == "" {
		logger.Fatal("KAFKA_TOPIC not set in configuration")
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker},
		GroupID:  "pulse-consumers",
		Topic:    topic,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	defer r.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := storage.Connect(ctx)
	if err != nil {
		logger.Fatal("ClickHouse connection error", zap.Error(err))
	}

	// Use mutex to synchronize access to the connection
	var mu sync.Mutex
	var processed, errors int

	logger.Info("Starting to consume messages",
		zap.String("broker", broker),
		zap.String("topic", topic))

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			logger.Error("Failed to read message from Kafka", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		var event models.Event
		if err := json.Unmarshal(m.Value, &event); err != nil {
			logger.Warn("Failed to unmarshal message",
				zap.Error(err),
				zap.String("payload", string(m.Value)))
			errors++
			continue
		}

		// Use a goroutine with mutex to handle concurrent writes safely
		go func(e models.Event) {
			mu.Lock()
			defer mu.Unlock()

			if err := storage.InsertEvent(ctx, conn, e); err != nil {
				logger.Error("Failed to insert event to ClickHouse",
					zap.Error(err),
					zap.String("service", e.Service),
					zap.Uint64("timestamp", e.EventTimeMs))
				errors++
				return
			}

			processed++
			if processed%1000 == 0 {
				logger.Info("Processing events",
					zap.Int("processed", processed),
					zap.Int("errors", errors))
			}
		}(event)
	}
}
