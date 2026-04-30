package natsinfra

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/yourusername/traffic-simulator/processor/internal/domain"
)

type ProcessedPublisher struct {
	js nats.JetStreamContext
}

func NewProcessedPublisher(js nats.JetStreamContext) *ProcessedPublisher {
	return &ProcessedPublisher{js: js}
}

func (p *ProcessedPublisher) PublishProcessed(ctx context.Context, msg domain.MessageEnvelope) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = p.js.Publish("messages.processed", data)
	return err
}
