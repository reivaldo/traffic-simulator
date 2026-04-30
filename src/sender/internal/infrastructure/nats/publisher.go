package natsinfra

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/yourusername/traffic-simulator/sender/internal/domain"
)

type DeliveryPublisher struct {
	js nats.JetStreamContext
}

func NewDeliveryPublisher(js nats.JetStreamContext) *DeliveryPublisher {
	return &DeliveryPublisher{js: js}
}

func (p *DeliveryPublisher) PublishSent(ctx context.Context, msg domain.MessageProcessed) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = p.js.Publish("messages.sent", data)
	return err
}
