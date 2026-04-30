package httpiface

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRouter_Healthz(t *testing.T) {
	t.Parallel()

	srv := NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
