package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	mu                     sync.Mutex
	isRunning              int32 = 0
	totalMessagesGenerated int64 = 0
	runCtx                 context.Context
	runCancel              context.CancelFunc
	ingestorURL            string
	httpClient                   = &http.Client{Timeout: 5 * time.Second}
	startTime              int64 = 0
	currentPhase           int32 = 0
	currentRate            int32 = 0
	totalDuration          int32 = 0
	workerPool             *WorkerPool

	// Prometheus metrics
	messagesGenerated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "simulator_messages_generated_total",
			Help: "Total number of messages generated",
		},
		[]string{"channel"},
	)
	currentMessageRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "simulator_current_rate",
			Help: "Current message generation rate (msg/sec)",
		},
		[]string{"phase"},
	)
	simulationRunning = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "simulator_running",
			Help: "Whether simulator is currently running (0 or 1)",
		},
	)
)

const (
	PHASE_BASELINE = iota
	PHASE_RAMP_UP
	PHASE_PEAK
	PHASE_RECOVERY
	PHASE_COOLDOWN
)

// WorkerPool manages N parallel goroutines for message delivery.
type WorkerPool struct {
	workers      int
	queue        chan Message
	wg           sync.WaitGroup
	sentCounter  int64
	errorCounter int64
}

func NewWorkerPool(numWorkers int) *WorkerPool {
	return &WorkerPool{
		workers: numWorkers,
		queue:   make(chan Message, numWorkers*2), // buffer = 2x workers
	}
}

// Start boots N worker goroutines.
func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}
}

// worker processes messages from the queue.
func (wp *WorkerPool) worker(ctx context.Context, id int) {
	defer wp.wg.Done()

	for {
		select {
		case <-ctx.Done():
			logrus.Debugf("Worker %d stopping", id)
			return
		case msg, ok := <-wp.queue:
			if !ok {
				return
			}
			// Sending is non-blocking for the producer side of the pool.
			if err := sendMessageToIngestor(ctx, msg); err != nil {
				atomic.AddInt64(&wp.errorCounter, 1)
				logrus.Debugf("Worker %d send failed: %v", id, err)
			} else {
				atomic.AddInt64(&wp.sentCounter, 1)
			}
		}
	}
}

// Submit adds a message to the queue.
func (wp *WorkerPool) Submit(ctx context.Context, msg Message) error {
	select {
	case wp.queue <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Stop waits for all workers to finish.
func (wp *WorkerPool) Stop() {
	close(wp.queue)
	wp.wg.Wait()
	logrus.Infof("Worker pool stopped. Sent: %d, Errors: %d",
		atomic.LoadInt64(&wp.sentCounter),
		atomic.LoadInt64(&wp.errorCounter))
}

type Message struct {
	ID        string    `json:"id"`
	Channel   string    `json:"channel"`
	Recipient string    `json:"recipient"`
	Subject   string    `json:"subject"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type SimulatorState struct {
	IsRunning      bool
	TotalGenerated int64
	CurrentPhase   int
	CurrentRate    int
	ElapsedSeconds int
	TotalDuration  int
}

func generateMessage(channel string) Message {
	channels := []string{"email", "sms", "whatsapp", "ota"}
	weights := []int{35, 30, 25, 10} // percentages for each channel

	if channel == "" {
		// Select channel based on weights
		randVal := rand.Intn(100)
		cumulative := 0
		for i, weight := range weights {
			cumulative += weight
			if randVal < cumulative {
				channel = channels[i]
				break
			}
		}
	}

	return Message{
		ID:        fmt.Sprintf("msg-%d-%d", time.Now().UnixNano(), rand.Int63()),
		Channel:   channel,
		Recipient: fmt.Sprintf("user%d@example.com", rand.Intn(10000)),
		Subject:   fmt.Sprintf("Test Message %d", rand.Intn(1000)),
		Content:   fmt.Sprintf("This is a test message at %s", time.Now().Format(time.RFC3339)),
		Timestamp: time.Now(),
	}
}

func sendMessageToIngestor(ctx context.Context, msg Message) error {
	payload := map[string]any{
		"external_id": msg.ID,
		"channel":     msg.Channel,
		"recipient":   msg.Recipient,
		"payload": map[string]any{
			"subject":   msg.Subject,
			"content":   msg.Content,
			"timestamp": msg.Timestamp.Format(time.RFC3339),
		},
	}
	jsonData, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ingestorURL+"/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logrus.Errorf("Failed to send message to ingestor: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		logrus.Warnf("Ingestor returned status %d", resp.StatusCode)
	}
	return nil
}

func escalationPattern(ctx context.Context, durationSeconds int) {
	logrus.Infof("Starting escalation pattern for %d seconds", durationSeconds)

	// Create and start a worker pool with 10 workers.
	workerPool = NewWorkerPool(10)
	workerPool.Start(ctx)
	defer workerPool.Stop()

	atomic.StoreInt64(&startTime, time.Now().Unix())
	atomic.StoreInt32(&totalDuration, int32(durationSeconds))

	// Realistic profile:
	// 1) ramp-up until 1:00
	// 2) sustained peak until 1:30
	// 3) controlled reduction afterwards, never dropping to zero
	rampDuration := 60
	peakEnd := 90
	if durationSeconds < rampDuration {
		rampDuration = durationSeconds
	}
	if durationSeconds < peakEnd {
		peakEnd = durationSeconds
	}

	baseRate := 80
	peakRate := 320
	residualRate := 140

	lastPhase := int32(-1)
	totalSent := 0

	runStartedAt := time.Now()
	for {
		elapsed := int(time.Since(runStartedAt).Seconds())
		if elapsed >= durationSeconds {
			break
		}

		if !atomic.CompareAndSwapInt32(&isRunning, 1, 1) {
			logrus.Info("Traffic generation stopped")
			simulationRunning.Set(0)
			return
		}

		select {
		case <-ctx.Done():
			simulationRunning.Set(0)
			return
		default:
		}

		rate := baseRate
		phaseName := "BASELINE"
		phaseID := int32(PHASE_BASELINE)

		switch {
		case elapsed < rampDuration && rampDuration > 0:
			progress := float64(elapsed) / float64(rampDuration)
			rate = baseRate + int(progress*float64(peakRate-baseRate))
			phaseName = "RAMP_UP"
			phaseID = PHASE_RAMP_UP
		case elapsed < peakEnd:
			rate = peakRate
			phaseName = "PEAK"
			phaseID = PHASE_PEAK
		default:
			tailDuration := durationSeconds - peakEnd
			if tailDuration <= 0 {
				rate = peakRate
			} else {
				progress := float64(elapsed-peakEnd) / float64(tailDuration)
				rate = peakRate - int(progress*float64(peakRate-residualRate))
				if rate < residualRate {
					rate = residualRate
				}
			}
			phaseName = "RECOVERY"
			phaseID = PHASE_RECOVERY
		}

		if phaseID != lastPhase {
			logrus.Infof("Entering %s phase at t=%ds (target=%d msg/sec)", phaseName, elapsed, rate)
			lastPhase = phaseID
		}

		atomic.StoreInt32(&currentPhase, phaseID)
		atomic.StoreInt32(&currentRate, int32(rate))
		currentMessageRate.WithLabelValues(phaseName).Set(float64(rate))

		// Send messages via worker pool (non-blocking producer path).
		secondSent := 0
		for i := 0; i < rate; i++ {
			select {
			case <-ctx.Done():
				simulationRunning.Set(0)
				return
			default:
				msg := generateMessage("")
				if err := workerPool.Submit(ctx, msg); err == nil {
					atomic.AddInt64(&totalMessagesGenerated, 1)
					messagesGenerated.WithLabelValues(msg.Channel).Inc()
					secondSent++
				}
			}
		}
		totalSent += secondSent
		time.Sleep(1 * time.Second)
	}

	atomic.StoreInt32(&isRunning, 0)
	simulationRunning.Set(0)
	logrus.Info("Escalation pattern completed")
	logrus.Infof("Total messages generated in run: %d", totalSent)
}

func init() {
	// Register Prometheus metrics
	prometheus.MustRegister(messagesGenerated)
	prometheus.MustRegister(currentMessageRate)
	prometheus.MustRegister(simulationRunning)
}

func main() {
	// Config loader
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	_ = viper.ReadInConfig()

	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	ingestorURL = viper.GetString("ingestor_url")
	if ingestorURL == "" {
		ingestorURL = "http://ingestor:8082"
	}

	logrus.Infof("Ingestor URL: %s", ingestorURL)

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	http.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ready")
	})

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		if !atomic.CompareAndSwapInt32(&isRunning, 0, 1) {
			http.Error(w, "simulation already running", http.StatusConflict)
			return
		}

		// Reset counters
		atomic.StoreInt64(&totalMessagesGenerated, 0)
		atomic.StoreInt32(&currentPhase, PHASE_BASELINE)
		atomic.StoreInt32(&currentRate, 0)
		simulationRunning.Set(1)

		duration := 120 // Default 2 minutes
		if d := r.URL.Query().Get("duration"); d != "" {
			fmt.Sscanf(d, "%d", &duration)
		}

		logrus.Infof("Starting traffic simulation for %d seconds", duration)
		mu.Lock()
		if runCancel != nil {
			runCancel()
		}
		runCtx, runCancel = context.WithCancel(context.Background())
		currentRunCtx := runCtx
		mu.Unlock()
		go escalationPattern(currentRunCtx, duration)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "started",
			"duration": duration,
		})
	})

	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		atomic.StoreInt32(&isRunning, 0)
		atomic.StoreInt32(&currentPhase, 0)
		atomic.StoreInt32(&currentRate, 0)
		simulationRunning.Set(0)
		mu.Lock()
		if runCancel != nil {
			runCancel()
			runCancel = nil
			runCtx = nil
		}
		mu.Unlock()

		logrus.Info("Traffic simulation stopped")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "stopped",
			"generated": atomic.LoadInt64(&totalMessagesGenerated),
		})
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		isRunningVal := atomic.LoadInt32(&isRunning) == 1
		elapsedSeconds := 0
		if isRunningVal {
			elapsedSeconds = int(time.Now().Unix() - atomic.LoadInt64(&startTime))
		}

		state := SimulatorState{
			IsRunning:      isRunningVal,
			TotalGenerated: atomic.LoadInt64(&totalMessagesGenerated),
			CurrentPhase:   int(atomic.LoadInt32(&currentPhase)),
			CurrentRate:    int(atomic.LoadInt32(&currentRate)),
			ElapsedSeconds: elapsedSeconds,
			TotalDuration:  int(atomic.LoadInt32(&totalDuration)),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(state)
	})

	port := viper.GetString("port")
	if port == "" {
		port = "8081"
	}

	logrus.Infof("Simulator listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
