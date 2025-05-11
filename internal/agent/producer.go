package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"os"

	"github.com/mohammadhptp/pulse/pkg/models"
	"github.com/segmentio/kafka-go"
)

func ProduceLogs(ctx context.Context, writer *kafka.Writer, input *os.File) error {
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		var event models.Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}
		msg, _ := json.Marshal(event)
		if err := writer.WriteMessages(ctx, kafka.Message{Value: msg}); err != nil {
			return err
		}
	}
	return scanner.Err()
}
