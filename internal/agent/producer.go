package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"os"

	"github.com/mohammadhptp/pulse/pkg/logger"
	"github.com/mohammadhptp/pulse/pkg/models"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

func ProduceLogs(ctx context.Context, writer *kafka.Writer, input *os.File) error {
	scanner := bufio.NewScanner(input)
	var processed, errors int
	
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			logger.Info("Stopping log processing due to context cancellation")
			return ctx.Err()
		default:
			var event models.Event
			if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
				logger.Warn("Failed to parse log entry", 
					zap.Error(err),
					zap.String("input", string(scanner.Bytes())))
				errors++
				continue
			}
			
			msg, err := json.Marshal(event)
			if err != nil {
				logger.Error("Failed to marshal event", zap.Error(err))
				errors++
				continue
			}
			
			if err := writer.WriteMessages(ctx, kafka.Message{Value: msg}); err != nil {
				logger.Error("Failed to write to Kafka", 
					zap.Error(err),
					zap.String("service", event.Service))
				return err
			}
			
			processed++
			if processed%1000 == 0 {
				logger.Info("Processing logs", 
					zap.Int("processed", processed), 
					zap.Int("errors", errors))
			}
		}
	}
	
	logger.Info("Completed log processing", 
		zap.Int("processed", processed), 
		zap.Int("errors", errors))
		
	return scanner.Err()
}
