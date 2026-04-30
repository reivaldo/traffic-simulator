package httpiface

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/traffic-simulator/ingestor/internal/application"
	"github.com/yourusername/traffic-simulator/ingestor/internal/domain"
)

type Handler struct {
	acceptUseCase *application.AcceptMessageIntent
	received      *prometheus.CounterVec
	published     *prometheus.CounterVec
}

func NewHandler(acceptUseCase *application.AcceptMessageIntent) *Handler {
	received := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "ingestor_messages_received_total", Help: "Total messages received"},
		[]string{"status"},
	)
	published := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "ingestor_messages_published_total", Help: "Total messages published to NATS"},
		[]string{"topic"},
	)
	prometheus.MustRegister(received, published)
	return &Handler{
		acceptUseCase: acceptUseCase,
		received:      received,
		published:     published,
	}
}

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok\n")) })
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ready\n")) })
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/messages", h.handleMessages)
	return mux
}

func (h *Handler) handleMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		h.received.WithLabelValues("error").Inc()
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	var req struct {
		ExternalID string         `json:"external_id"`
		Channel    string         `json:"channel"`
		Recipient  string         `json:"recipient"`
		Payload    map[string]any `json:"payload"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		h.received.WithLabelValues("error").Inc()
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		http.Error(w, "invalid json", http.StatusBadRequest)
		h.received.WithLabelValues("error").Inc()
		return
	}
	intent, err := domain.NewMessageIntent(req.ExternalID, req.Channel, req.Recipient, req.Payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		h.received.WithLabelValues("error").Inc()
		return
	}
	if err := h.acceptUseCase.Execute(r.Context(), intent); err != nil {
		logrus.Errorf("publish accepted message failed: %v", err)
		http.Error(w, "nats publish failed", http.StatusInternalServerError)
		h.received.WithLabelValues("error").Inc()
		return
	}
	h.received.WithLabelValues("success").Inc()
	h.published.WithLabelValues("messages.incoming").Inc()
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte("published\n"))
}
