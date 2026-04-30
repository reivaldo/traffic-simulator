package application

import (
	"context"

	"github.com/yourusername/traffic-simulator/processor/internal/domain"
)

// ProcessAcceptedMessagePort is the contract used by the JetStream subscriber to run business logic.
type ProcessAcceptedMessagePort interface {
	Execute(ctx context.Context, msg domain.MessageEnvelope) error
}

type MessageRepositoryPort interface {
	SaveProcessed(ctx context.Context, msg domain.MessageEnvelope) error
	MarkPublished(ctx context.Context, externalID string) error
}

type ProcessedEventPublisherPort interface {
	PublishProcessed(ctx context.Context, msg domain.MessageEnvelope) error
}
