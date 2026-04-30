package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var simulatorURL string
var simulatorClient = &http.Client{Timeout: 5 * time.Second}

var (
	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "admin_api_requests_total",
			Help: "Total HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "admin_api_request_duration_seconds",
			Help: "HTTP request duration in seconds",
		},
		[]string{"method", "path"},
	)
)

func init() {
	prometheus.MustRegister(httpRequests)
	prometheus.MustRegister(httpDuration)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		httpRequests.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(rec.status)).Inc()
		httpDuration.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(start).Seconds())
	})
}

func doSimulatorGet(path string) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, simulatorURL+path, nil)
	if err != nil {
		return nil, err
	}
	return simulatorClient.Do(req)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	_ = viper.ReadInConfig()

	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	simulatorURL = viper.GetString("simulator_url")
	if simulatorURL == "" {
		simulatorURL = "http://simulator:8081"
	}

	logrus.Infof("Simulator URL: %s", simulatorURL)

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ready")
	})

	// Metrics endpoints
	mux.HandleFunc("/metrics/simulator", func(w http.ResponseWriter, r *http.Request) {
		resp, err := doSimulatorGet("/status")
		if err != nil {
			respondJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "simulator unavailable"})
			return
		}
		defer resp.Body.Close()

		var status map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&status)

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status": "success",
			"data":   status,
		})
	})

	// Control endpoints
	mux.HandleFunc("/simulator/start", func(w http.ResponseWriter, r *http.Request) {
		duration := r.URL.Query().Get("duration")
		if duration == "" {
			duration = "300"
		}

		resp, err := doSimulatorGet("/start?duration=" + duration)
		if err != nil {
			respondJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "failed to start simulator"})
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status": "success",
			"data":   data,
		})
	})

	mux.HandleFunc("/simulator/stop", func(w http.ResponseWriter, r *http.Request) {
		resp, err := doSimulatorGet("/stop")
		if err != nil {
			respondJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "failed to stop simulator"})
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status": "success",
			"data":   data,
		})
	})

	mux.HandleFunc("/simulator/status", func(w http.ResponseWriter, r *http.Request) {
		resp, err := doSimulatorGet("/status")
		if err != nil {
			respondJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "simulator unavailable"})
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status": "success",
			"data":   data,
		})
	})

	// Add Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	port := viper.GetString("port")
	if port == "" {
		port = "8086"
	}

	logrus.Infof("Admin API listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, corsMiddleware(metricsMiddleware(mux))))
}
