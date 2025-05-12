package agent

import (
	"context"
	"encoding/json"

	"github.com/mohammadhptp/pulse/pkg/logger"
	"github.com/mohammadhptp/pulse/pkg/models"
	"github.com/mohammadhptp/pulse/pkg/transport"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type EventProcessor struct {
	writer    *kafka.Writer
	transport transport.EventProducer
	processed int
	errors    int
}

func NewEventProcessor(writer *kafka.Writer, t transport.EventProducer) *EventProcessor {
	processor := &EventProcessor{
		writer:    writer,
		transport: t,
		processed: 0,
		errors:    0,
	}

	t.SetEventHandler(processor.handleEvent)

	return processor
}

func (p *EventProcessor) Start(ctx context.Context) error {
	logger.Info("Starting event processor")

	if err := p.transport.Start(ctx); err != nil {
		return err
	}

	<-ctx.Done()

	logger.Info("Completed event processing",
		zap.Int("processed", p.processed),
		zap.Int("errors", p.errors))

	return p.transport.Close()
}

func (p *EventProcessor) handleEvent(event models.Event) error {
	msg, err := json.Marshal(event)
	if err != nil {
		logger.Error("Failed to marshal event", zap.Error(err))
		p.errors++
		return err
	}

	ctx := context.Background()
	if err := p.writer.WriteMessages(ctx, kafka.Message{Value: msg}); err != nil {
		logger.Error("Failed to write to Kafka",
			zap.Error(err),
			zap.String("service", event.Service))
		p.errors++
		return err
	}

	p.processed++
	if p.processed%1000 == 0 {
		logger.Info("Processing events",
			zap.Int("processed", p.processed),
			zap.Int("errors", p.errors))
	}

	return nil
}

func ProduceLogs(ctx context.Context, writer *kafka.Writer, input interface{}) error {
	logger.Warn("ProduceLogs is deprecated, please use EventProcessor instead")

	return ctx.Err()
}
