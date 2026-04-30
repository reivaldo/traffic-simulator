package application

import (
	"context"

	"github.com/yourusername/traffic-simulator/notification-service/internal/domain"
)

// ProviderGateway defines the contract for delivery providers (SMS, Email, WhatsApp, etc.)
type ProviderGateway interface {
	Send(ctx context.Context, notification domain.Notification) (status string, err error)
}
