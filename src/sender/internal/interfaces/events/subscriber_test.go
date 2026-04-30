package eventsiface

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourusername/traffic-simulator/sender/internal/domain"
)

type mockDeliver struct {
	executeFn         func(ctx context.Context, msg domain.MessageProcessed) error
	maxConcurrentSeen int32
	current           int32
}

func (m *mockDeliver) Execute(ctx context.Context, msg domain.MessageProcessed) error {
	cur := atomic.AddInt32(&m.current, 1)
	for {
		old := atomic.LoadInt32(&m.maxConcurrentSeen)
		if cur <= old {
			break
		}
		if atomic.CompareAndSwapInt32(&m.maxConcurrentSeen, old, cur) {
			break
		}
	}
	time.Sleep(20 * time.Millisecond)
	atomic.AddInt32(&m.current, -1)
	if m.executeFn != nil {
		return m.executeFn(ctx, msg)
	}
	return nil
}

func TestSubscriber_SemaphoreRespectsMaxConcurrent(t *testing.T) {
	t.Parallel()

	mock := &mockDeliver{}
	sub := NewSubscriber(mock, SubscriberOptions{MaxConcurrentMessages: 5})

	for range 50 {
		go func() {
			sub.sem <- struct{}{}
			go func() {
				defer func() { <-sub.sem }()
				_ = mock.Execute(context.Background(), domain.MessageProcessed{ExternalID: "x"})
			}()
		}()
	}
	time.Sleep(150 * time.Millisecond)
	max := atomic.LoadInt32(&mock.maxConcurrentSeen)
	if max > 5 {
		t.Fatalf("expected at most 5 concurrent, got %d", max)
	}
}

func TestSubscriber_DefaultMaxConcurrent50(t *testing.T) {
	t.Parallel()

	sub := NewSubscriber(&mockDeliver{}, SubscriberOptions{MaxConcurrentMessages: 0})
	if cap(sub.sem) != 50 {
		t.Fatalf("expected default cap 50, got %d", cap(sub.sem))
	}
}

func TestSubscriber_ConfigurableMax(t *testing.T) {
	t.Parallel()

	sub := NewSubscriber(&mockDeliver{}, SubscriberOptions{MaxConcurrentMessages: 12})
	if cap(sub.sem) != 12 {
		t.Fatalf("expected cap 12, got %d", cap(sub.sem))
	}
}
