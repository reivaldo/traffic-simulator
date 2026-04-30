package application

import (
	"context"

	"github.com/yourusername/traffic-simulator/ingestor/internal/domain"
)

type AcceptMessageIntent struct {
	publisher IntentPublisherPort
}

func NewAcceptMessageIntent(publisher IntentPublisherPort) *AcceptMessageIntent {
	return &AcceptMessageIntent{publisher: publisher}
}

func (uc *AcceptMessageIntent) Execute(ctx context.Context, intent domain.MessageIntent) error {
	return uc.publisher.PublishAccepted(ctx, intent)
}
