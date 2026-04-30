package application

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/traffic-simulator/notification-service/internal/domain"
)

// SendToAllProviders orchestrates parallel delivery to multiple providers (fan-out pattern)
type SendToAllProviders struct {
	gateways map[string]ProviderGateway
}

func NewSendToAllProviders(gateways map[string]ProviderGateway) *SendToAllProviders {
	return &SendToAllProviders{gateways: gateways}
}

// Execute sends notification to all providers in parallel
// Returns results from each provider
// Success = at least one provider succeeds
func (uc *SendToAllProviders) Execute(ctx context.Context, notification domain.Notification) ([]domain.ProviderResult, error) {
	if len(uc.gateways) == 0 {
		return nil, fmt.Errorf("no providers configured")
	}

	results := make([]domain.ProviderResult, 0)
	resultChan := make(chan domain.ProviderResult, len(uc.gateways))
	var wg sync.WaitGroup

	// Fan-out: spawn 1 goroutine per provider
	for providerName, gateway := range uc.gateways {
		wg.Add(1)
		go func(name string, gw ProviderGateway) {
			defer wg.Done()

			start := time.Now()
			status, err := gw.Send(ctx, notification)
			duration := time.Since(start).Seconds() * 1000 // milliseconds

			logrus.Debugf("Provider %s completed in %.2fms: status=%s, err=%v", name, duration, status, err)

			resultChan <- domain.ProviderResult{
				Provider: name,
				Status:   status,
				Error:    err,
				Duration: duration,
			}
		}(providerName, gateway)
	}

	// Close channel when all goroutines finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for result := range resultChan {
		results = append(results, result)
	}

	// Determine outcome
	successCount := 0
	var failedProviders []string
	for _, result := range results {
		if result.Error == nil {
			successCount++
		} else {
			failedProviders = append(failedProviders, result.Provider)
		}
	}

	// Log summary
	logrus.Infof("Notification delivery [external_id=%s, channel=%s]: %d/%d providers succeeded",
		notification.ExternalID, notification.Channel, successCount, len(uc.gateways))

	if len(failedProviders) > 0 {
		logrus.Warnf("Providers failed: %v", failedProviders)
	}

	// Success if at least one provider succeeds (resilient)
	if successCount == 0 {
		return results, fmt.Errorf("all providers failed")
	}

	return results, nil
}
