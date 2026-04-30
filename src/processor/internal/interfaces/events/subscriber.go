package eventsiface

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/traffic-simulator/processor/internal/application"
	"github.com/yourusername/traffic-simulator/processor/internal/domain"
)

var (
	processorMessagesProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "processor_messages_processed_total",
			Help: "Total messages processed by processor",
		},
		[]string{"status"},
	)
	processorDBOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "processor_db_operations_total",
			Help: "Total processor database operations",
		},
		[]string{"operation", "status"},
	)
	processorConcurrentMessages = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "processor_messages_concurrent",
			Help: "Current number of messages being processed concurrently",
		},
	)
)

func init() {
	prometheus.MustRegister(
		processorMessagesProcessed,
		processorDBOperations,
		processorConcurrentMessages,
	)
}

// SubscriberOptions configures semaphore limits
type SubscriberOptions struct {
	MaxConcurrentMessages int
}

type Subscriber struct {
	useCase application.ProcessAcceptedMessagePort
	sem     chan struct{} // Semaphore: buffered channel with capacity = max concurrent
}

func NewSubscriber(useCase application.ProcessAcceptedMessagePort, opts SubscriberOptions) *Subscriber {
	maxConcurrent := opts.MaxConcurrentMessages
	if maxConcurrent <= 0 {
		maxConcurrent = 50 // default
	}
	
	logrus.Infof("Processor semaphore configured: max %d concurrent messages", maxConcurrent)
	
	return &Subscriber{
		useCase: useCase,
		sem:     make(chan struct{}, maxConcurrent),
	}
}

func (s *Subscriber) Subscribe(js nats.JetStreamContext) error {
	_, err := js.Subscribe("messages.incoming", func(msg *nats.Msg) {
		// Semaphore: acquire slot (blocks if at capacity)
		select {
		case s.sem <- struct{}{}:
			// Slot acquired, process asynchronously
			go s.processMessageAsync(msg)
		case <-time.After(100 * time.Millisecond):
			// Timeout acquiring slot (shouldn't happen with proper sizing)
			logrus.Warnf("Timeout acquiring semaphore slot, rejecting message")
			_ = msg.Nak()
		}
	}, nats.Durable("processor"), nats.ManualAck())
	return err
}

// processMessageAsync handles the actual message processing with semaphore slot acquired
func (s *Subscriber) processMessageAsync(msg *nats.Msg) {
	// Release semaphore slot when done (guaranteed)
	defer func() { <-s.sem }()
	
	// Update metrics
	processorConcurrentMessages.Inc()
	defer processorConcurrentMessages.Dec()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var envelope domain.MessageEnvelope
	if err := json.Unmarshal(msg.Data, &envelope); err != nil {
		logrus.Errorf("invalid message payload: %v", err)
		processorMessagesProcessed.WithLabelValues("error").Inc()
		_ = msg.Ack()
		return
	}

	if err := s.useCase.Execute(ctx, envelope); err != nil {
		logrus.Errorf("processing failed for external_id=%s: %v", envelope.ExternalID, err)
		processorMessagesProcessed.WithLabelValues("error").Inc()
		processorDBOperations.WithLabelValues("insert", "error").Inc()
		_ = msg.Nak()
		return
	}

	logrus.Debugf("Processing message: external_id=%s, channel=%s", envelope.ExternalID, envelope.Channel)
	processorMessagesProcessed.WithLabelValues("success").Inc()
	processorDBOperations.WithLabelValues("insert", "success").Inc()
	_ = msg.Ack()
}
