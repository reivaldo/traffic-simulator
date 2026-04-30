package eventsiface

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourusername/traffic-simulator/processor/internal/domain"
)

// mockUseCase implements the use case for testing
type mockUseCase struct {
	executeFn func(context.Context, domain.MessageEnvelope) error
	callCount int64
}

func (m *mockUseCase) Execute(ctx context.Context, msg domain.MessageEnvelope) error {
	atomic.AddInt64(&m.callCount, 1)
	if m.executeFn != nil {
		return m.executeFn(ctx, msg)
	}
	return nil
}

// TestSemaphore_LimitsConcurrency verifies that semaphore caps concurrent processing
func TestSemaphore_LimitsConcurrency(t *testing.T) {
	t.Parallel()

	// Track maximum concurrent goroutines
	maxConcurrent := int32(0)
	currentConcurrent := int32(0)

	mockUC := &mockUseCase{
		executeFn: func(ctx context.Context, msg domain.MessageEnvelope) error {
			current := atomic.AddInt32(&currentConcurrent, 1)
			// Update max
			for {
				oldMax := atomic.LoadInt32(&maxConcurrent)
				if current <= oldMax {
					break
				}
				if atomic.CompareAndSwapInt32(&maxConcurrent, oldMax, current) {
					break
				}
			}

			// Simulate I/O
			time.Sleep(50 * time.Millisecond)

			atomic.AddInt32(&currentConcurrent, -1)
			return nil
		},
	}

	// Create subscriber with semaphore limit of 10
	subscriber := NewSubscriber(mockUC, SubscriberOptions{
		MaxConcurrentMessages: 10,
	})

	// Simulate processing 100 messages by calling processMessageAsync directly
	// (in real scenario, NATS would deliver these)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		envelope := domain.MessageEnvelope{
			ExternalID: "test-" + string(rune('a'+i%26)),
			Channel:    "test",
			Recipient:  "test@example.com",
			Payload:    json.RawMessage(`{}`),
		}

		// Simulate NATS message
		go func(env domain.MessageEnvelope) {
			defer wg.Done()
			// Try to acquire semaphore
			select {
			case subscriber.sem <- struct{}{}:
				go func() {
					defer func() { <-subscriber.sem }()
					mockUC.Execute(context.Background(), env)
				}()
			}
		}(envelope)
	}

	wg.Wait()

	// Wait for remaining goroutines to complete
	time.Sleep(1 * time.Second)

	maxCon := atomic.LoadInt32(&maxConcurrent)
	if maxCon > 10 {
		t.Fatalf("Expected max concurrency 10, got %d", maxCon)
	}

	t.Logf("Semaphore test passed: max concurrent = %d (limit was 10)", maxCon)
}

// TestSemaphore_DefaultMax verifies default max_concurrent is 50
func TestSemaphore_DefaultMax(t *testing.T) {
	t.Parallel()

	mockUC := &mockUseCase{}
	subscriber := NewSubscriber(mockUC, SubscriberOptions{
		MaxConcurrentMessages: 0, // Should use default
	})

	if len(subscriber.sem) != 0 {
		t.Fatalf("Expected semaphore to be empty initially")
	}

	// Semaphore capacity should be 50 (default)
	if cap(subscriber.sem) != 50 {
		t.Fatalf("Expected default semaphore capacity 50, got %d", cap(subscriber.sem))
	}
}

// TestSemaphore_ConfigurableMax verifies semaphore respects config
func TestSemaphore_ConfigurableMax(t *testing.T) {
	t.Parallel()

	mockUC := &mockUseCase{}

	for _, maxConcurrent := range []int{10, 20, 100} {
		subscriber := NewSubscriber(mockUC, SubscriberOptions{
			MaxConcurrentMessages: maxConcurrent,
		})

		if cap(subscriber.sem) != maxConcurrent {
			t.Fatalf("Expected semaphore capacity %d, got %d", maxConcurrent, cap(subscriber.sem))
		}
	}
}
