package collector

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/mohammadhptp/pulse/internal/storage"
	"github.com/mohammadhptp/pulse/pkg/models"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
)

func Run() {
	broker := viper.GetString("KAFKA_BROKER")
	topic := viper.GetString("KAFKA_TOPIC")

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker},
		GroupID:  "pulse-consumers",
		Topic:    topic,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	defer r.Close()

	ctx := context.Background()
	conn, err := storage.Connect(ctx)
	if err != nil {
		log.Fatalf("ClickHouse conn error: %v", err)
	}

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			log.Printf("Read error: %v", err)
			time.Sleep(time.Second)
			continue
		}
		var event models.Event
		if err := json.Unmarshal(m.Value, &event); err != nil {
			continue
		}
		go storage.InsertEvent(ctx, conn, event)
	}
}
