package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/traffic-simulator/notification-service/internal/domain"
)

func TestSMSGateway_Send_Success(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/send" {
			t.Errorf("path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer ts.Close()

	g := NewSMSGateway(ts.URL)
	st, err := g.Send(context.Background(), domain.Notification{
		ExternalID: "x",
		Channel:    "sms",
		Recipient:  "n",
	})
	if err != nil {
		t.Fatal(err)
	}
	if st != "ok" {
		t.Fatalf("status %q", st)
	}
}

func TestEmailGateway_Send_SuccessString(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "delivered"})
	}))
	defer ts.Close()

	g := NewEmailGateway(ts.URL)
	st, err := g.Send(context.Background(), domain.Notification{ExternalID: "e"})
	if err != nil {
		t.Fatal(err)
	}
	if st != "delivered" {
		t.Fatalf("got %q", st)
	}
}

func TestWhatsAppGateway_Send_FailsWhenNoStatusField(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"foo":1}`))
	}))
	defer ts.Close()

	g := NewWhatsAppGateway(ts.URL)
	st, err := g.Send(context.Background(), domain.Notification{ExternalID: "w"})
	if err == nil {
		t.Fatal("expected error when status field is missing")
	}
	if st != "failed" {
		t.Fatalf("expected failed status, got %q", st)
	}
}

func TestEmailGateway_Send_FailsOnNon2xx(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "error"})
	}))
	defer ts.Close()

	g := NewEmailGateway(ts.URL)
	st, err := g.Send(context.Background(), domain.Notification{ExternalID: "e-2"})
	if err == nil {
		t.Fatal("expected error for non-2xx response")
	}
	if st != "failed" {
		t.Fatalf("expected failed status, got %q", st)
	}
}
