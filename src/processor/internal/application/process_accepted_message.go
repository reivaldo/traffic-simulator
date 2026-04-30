package application

import (
	"context"

	"github.com/yourusername/traffic-simulator/processor/internal/domain"
)

type ProcessAcceptedMessage struct {
	repo      MessageRepositoryPort
	publisher ProcessedEventPublisherPort
}

func NewProcessAcceptedMessage(repo MessageRepositoryPort, publisher ProcessedEventPublisherPort) *ProcessAcceptedMessage {
	return &ProcessAcceptedMessage{repo: repo, publisher: publisher}
}

func (uc *ProcessAcceptedMessage) Execute(ctx context.Context, msg domain.MessageEnvelope) error {
	if err := uc.repo.SaveProcessed(ctx, msg); err != nil {
		return err
	}
	if err := uc.publisher.PublishProcessed(ctx, msg); err != nil {
		return err
	}
	return uc.repo.MarkPublished(ctx, msg.ExternalID)
}
