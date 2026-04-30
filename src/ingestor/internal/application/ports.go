package application

import (
	"context"

	"github.com/yourusername/traffic-simulator/ingestor/internal/domain"
)

type IntentPublisherPort interface {
	PublishAccepted(ctx context.Context, intent domain.MessageIntent) error
}
