package application

import (
	"context"
	"errors"
	"testing"

	"github.com/yourusername/traffic-simulator/notification-service/internal/domain"
)

// Fake gateway for testing
type fakeGateway struct {
	shouldFail bool
	callCount  int
}

func (f *fakeGateway) Send(ctx context.Context, notification domain.Notification) (status string, err error) {
	f.callCount++
	if f.shouldFail {
		return "", errors.New("simulated failure")
	}
	return "success", nil
}

func TestFanOut_AllSuccess(t *testing.T) {
	t.Parallel()

	gateways := map[string]ProviderGateway{
		"sms":      &fakeGateway{shouldFail: false},
		"email":    &fakeGateway{shouldFail: false},
		"whatsapp": &fakeGateway{shouldFail: false},
	}

	uc := NewSendToAllProviders(gateways)
	notification := domain.Notification{
		ExternalID: "msg-123",
		Channel:    "email",
		Recipient:  "user@example.com",
	}

	results, err := uc.Execute(context.Background(), notification)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	for _, result := range results {
		if result.Error != nil {
			t.Fatalf("expected no error in result, got %v", result.Error)
		}
		if result.Status != "success" {
			t.Fatalf("expected status 'success', got %s", result.Status)
		}
	}
}

func TestFanOut_PartialFailure(t *testing.T) {
	t.Parallel()

	gateways := map[string]ProviderGateway{
		"sms":      &fakeGateway{shouldFail: false},
		"email":    &fakeGateway{shouldFail: true}, // Fails
		"whatsapp": &fakeGateway{shouldFail: false},
	}

	uc := NewSendToAllProviders(gateways)
	notification := domain.Notification{
		ExternalID: "msg-456",
		Channel:    "sms",
		Recipient:  "1234567890",
	}

	results, err := uc.Execute(context.Background(), notification)

	if err != nil {
		t.Fatalf("expected no error (partial success ok), got %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	successCount := 0
	for _, result := range results {
		if result.Error == nil {
			successCount++
		}
	}

	if successCount != 2 {
		t.Fatalf("expected 2 successes, got %d", successCount)
	}
}

func TestFanOut_AllFailure(t *testing.T) {
	t.Parallel()

	gateways := map[string]ProviderGateway{
		"sms":      &fakeGateway{shouldFail: true},
		"email":    &fakeGateway{shouldFail: true},
		"whatsapp": &fakeGateway{shouldFail: true},
	}

	uc := NewSendToAllProviders(gateways)
	notification := domain.Notification{
		ExternalID: "msg-789",
		Channel:    "whatsapp",
		Recipient:  "+5511987654321",
	}

	results, err := uc.Execute(context.Background(), notification)

	if err == nil {
		t.Fatalf("expected error when all fail, got nil")
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	for _, result := range results {
		if result.Error == nil {
			t.Fatalf("expected error in all results")
		}
	}
}

func TestFanOut_ParallelExecution(t *testing.T) {
	t.Parallel()

	fake1 := &fakeGateway{shouldFail: false}
	fake2 := &fakeGateway{shouldFail: false}
	fake3 := &fakeGateway{shouldFail: false}

	gateways := map[string]ProviderGateway{
		"sms":      fake1,
		"email":    fake2,
		"whatsapp": fake3,
	}

	uc := NewSendToAllProviders(gateways)
	notification := domain.Notification{
		ExternalID: "msg-parallel",
		Channel:    "test",
		Recipient:  "test@example.com",
	}

	results, err := uc.Execute(context.Background(), notification)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// All should have been called exactly once
	if fake1.callCount != 1 || fake2.callCount != 1 || fake3.callCount != 1 {
		t.Fatalf("expected each gateway to be called once, got %d, %d, %d",
			fake1.callCount, fake2.callCount, fake3.callCount)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}
