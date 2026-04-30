package application

import (
	"context"
	"errors"
	"testing"

	"github.com/yourusername/traffic-simulator/sender/internal/domain"
)

type fakeDeliveryPublisher struct {
	err    error
	called bool
}

func (f *fakeDeliveryPublisher) PublishSent(_ context.Context, _ domain.MessageProcessed) error {
	f.called = true
	return f.err
}

func TestExecute_DelegatesToPublisher(t *testing.T) {
	t.Parallel()

	pub := &fakeDeliveryPublisher{}
	uc := NewDeliverProcessedMessage(pub)
	if err := uc.Execute(context.Background(), domain.MessageProcessed{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !pub.called {
		t.Fatal("expected publisher to be called")
	}
}

func TestExecute_PropagatesError(t *testing.T) {
	t.Parallel()

	pub := &fakeDeliveryPublisher{err: errors.New("downstream failure")}
	uc := NewDeliverProcessedMessage(pub)
	if err := uc.Execute(context.Background(), domain.MessageProcessed{}); err == nil {
		t.Fatal("expected error")
	}
}
