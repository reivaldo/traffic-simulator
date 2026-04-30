package application

import (
	"context"
	"errors"
	"testing"

	"github.com/yourusername/traffic-simulator/processor/internal/domain"
)

type fakeRepo struct {
	saveErr         error
	markPublishedErr error
	saved           bool
	published       bool
}

func (f *fakeRepo) SaveProcessed(_ context.Context, _ domain.MessageEnvelope) error {
	f.saved = true
	return f.saveErr
}

func (f *fakeRepo) MarkPublished(_ context.Context, _ string) error {
	f.published = true
	return f.markPublishedErr
}

type fakePublisher struct {
	err    error
	called bool
}

func (f *fakePublisher) PublishProcessed(_ context.Context, _ domain.MessageEnvelope) error {
	f.called = true
	return f.err
}

func TestExecute_MarksPublishedOnlyAfterPublish(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{}
	pub := &fakePublisher{}
	uc := NewProcessAcceptedMessage(repo, pub)

	err := uc.Execute(context.Background(), domain.MessageEnvelope{ExternalID: "ext-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.saved || !pub.called || !repo.published {
		t.Fatalf("expected save, publish and mark published to be called, got save=%v publish=%v mark=%v", repo.saved, pub.called, repo.published)
	}
}

func TestExecute_DoesNotMarkPublishedWhenPublishFails(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{}
	pub := &fakePublisher{err: errors.New("publish failed")}
	uc := NewProcessAcceptedMessage(repo, pub)

	err := uc.Execute(context.Background(), domain.MessageEnvelope{ExternalID: "ext-1"})
	if err == nil {
		t.Fatal("expected error")
	}
	if repo.published {
		t.Fatal("mark published should not be called on publish failure")
	}
}
