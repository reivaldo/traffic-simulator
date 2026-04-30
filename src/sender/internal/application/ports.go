package application

import (
	"context"

	"github.com/yourusername/traffic-simulator/sender/internal/domain"
)

// DeliverProcessedMessagePort is the contract used by the JetStream subscriber.
type DeliverProcessedMessagePort interface {
	Execute(ctx context.Context, msg domain.MessageProcessed) error
}

type DeliveryEventPublisherPort interface {
	PublishSent(ctx context.Context, msg domain.MessageProcessed) error
}
