package application

import (
	"context"

	"github.com/yourusername/traffic-simulator/sender/internal/domain"
)

type DeliverProcessedMessage struct {
	publisher DeliveryEventPublisherPort
}

func NewDeliverProcessedMessage(publisher DeliveryEventPublisherPort) *DeliverProcessedMessage {
	return &DeliverProcessedMessage{publisher: publisher}
}

func (uc *DeliverProcessedMessage) Execute(ctx context.Context, msg domain.MessageProcessed) error {
	return uc.publisher.PublishSent(ctx, msg)
}
