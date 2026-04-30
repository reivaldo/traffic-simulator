package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewWorkerPoolQueueCapacity(t *testing.T) {
	t.Parallel()

	wp := NewWorkerPool(7)
	if cap(wp.queue) != 14 {
		t.Fatalf("expected queue cap 14 (2x workers), got %d", cap(wp.queue))
	}
	if wp.workers != 7 {
		t.Fatalf("expected 7 workers")
	}
}

func TestGenerateMessage_FixedChannel(t *testing.T) {
	t.Parallel()

	m := generateMessage("sms")
	if m.Channel != "sms" {
		t.Fatalf("expected channel sms, got %q", m.Channel)
	}
}

func TestWorkerPool_HTTPSuccessIncrementsSent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/messages" {
			t.Errorf("path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	oldURL := ingestorURL
	ingestorURL = ts.URL
	defer func() { ingestorURL = oldURL }()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	wp := NewWorkerPool(2)
	wp.Start(ctx)
	defer wp.Stop()

	msg := Message{
		ID:        "id-1",
		Channel:   "email",
		Recipient: "a@b.c",
		Subject:   "s",
		Content:   "c",
		Timestamp: time.Now(),
	}
	if err := wp.Submit(ctx, msg); err != nil {
		t.Fatal(err)
	}

	// wait for worker to process
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt64(&wp.sentCounter) >= 1 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if atomic.LoadInt64(&wp.sentCounter) < 1 {
		t.Fatalf("expected at least one successful send, sent=%d err=%d",
			atomic.LoadInt64(&wp.sentCounter), atomic.LoadInt64(&wp.errorCounter))
	}
}

func TestMainPackage(t *testing.T) {}
