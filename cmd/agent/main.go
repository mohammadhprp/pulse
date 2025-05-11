package main

import (
	"context"
	"log"
	"os"

	"github.com/mohammadhptp/pulse/internal/agent"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Config error: %v", err)
	}

	broker := viper.GetString("KAFKA_BROKER")
	topic := viper.GetString("KAFKA_TOPIC")

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{broker},
		Topic:   topic,
	})
	defer writer.Close()

	if err := agent.ProduceLogs(context.Background(), writer, os.Stdin); err != nil {
		log.Fatalf("Produce error: %v", err)
	}
}
