package application

import (
	"context"
	"errors"
	"testing"

	"github.com/yourusername/traffic-simulator/ingestor/internal/domain"
)

type fakeIntentPublisher struct {
	called bool
	err    error
}

func (f *fakeIntentPublisher) PublishAccepted(_ context.Context, _ domain.MessageIntent) error {
	f.called = true
	return f.err
}

func TestAcceptMessageIntent_Execute_DelegatesToPublisher(t *testing.T) {
	t.Parallel()

	pub := &fakeIntentPublisher{}
	uc := NewAcceptMessageIntent(pub)
	intent, _ := domain.NewMessageIntent("e1", "email", "a@b.c", nil)

	if err := uc.Execute(context.Background(), intent); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !pub.called {
		t.Fatal("expected publisher to be called")
	}
}

func TestAcceptMessageIntent_Execute_PropagatesError(t *testing.T) {
	t.Parallel()

	pub := &fakeIntentPublisher{err: errors.New("nats down")}
	uc := NewAcceptMessageIntent(pub)
	intent, _ := domain.NewMessageIntent("e1", "email", "a@b.c", nil)

	if err := uc.Execute(context.Background(), intent); err == nil {
		t.Fatal("expected error")
	}
}
