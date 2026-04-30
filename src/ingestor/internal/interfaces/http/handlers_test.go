package httpiface

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/yourusername/traffic-simulator/ingestor/internal/application"
	"github.com/yourusername/traffic-simulator/ingestor/internal/domain"
)

var (
	handlerOnce sync.Once
	testHandler *Handler
)

type fakePublisher struct {
	called bool
	err    error
}

func (f *fakePublisher) PublishAccepted(_ context.Context, _ domain.MessageIntent) error {
	f.called = true
	return f.err
}

func TestHandleMessages_RejectsUnknownFields(t *testing.T) {
	pub := &fakePublisher{}
	handlerOnce.Do(func() {
		testHandler = NewHandler(application.NewAcceptMessageIntent(pub))
	})

	body := []byte(`{"external_id":"id-1","channel":"email","recipient":"a@b.com","payload":{"x":"y"},"unknown":"x"}`)
	req := httptest.NewRequest(http.MethodPost, "/messages", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	testHandler.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleMessages_AcceptsValidRequest(t *testing.T) {
	pub := &fakePublisher{}
	handlerOnce.Do(func() {
		testHandler = NewHandler(application.NewAcceptMessageIntent(pub))
	})

	body := []byte(`{"external_id":"id-1","channel":"email","recipient":"a@b.com","payload":{"x":"y"}}`)
	req := httptest.NewRequest(http.MethodPost, "/messages", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	testHandler.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}
}
