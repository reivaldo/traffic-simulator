package natsinfra

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/yourusername/traffic-simulator/ingestor/internal/domain"
)

type IntentPublisher struct {
	nc *nats.Conn
}

func NewIntentPublisher(nc *nats.Conn) *IntentPublisher {
	return &IntentPublisher{nc: nc}
}

func (p *IntentPublisher) PublishAccepted(ctx context.Context, intent domain.MessageIntent) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	data, err := json.Marshal(map[string]any{
		"external_id": intent.ExternalID,
		"channel":     intent.Channel,
		"recipient":   intent.Recipient,
		"payload":     intent.Payload,
	})
	if err != nil {
		return err
	}
	if err := p.nc.Publish("messages.incoming", data); err != nil {
		return err
	}
	return p.nc.Flush()
}
