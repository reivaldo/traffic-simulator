package eventsiface

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/traffic-simulator/sender/internal/application"
	"github.com/yourusername/traffic-simulator/sender/internal/domain"
)

var (
	senderMessagesSent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sender_messages_sent_total",
			Help: "Total messages sent by sender",
		},
		[]string{"status"},
	)
	senderConcurrentMessages = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "sender_messages_concurrent",
			Help: "Current number of messages being sent concurrently",
		},
	)
)

func init() {
	prometheus.MustRegister(senderMessagesSent, senderConcurrentMessages)
}

// SubscriberOptions configures semaphore limits
type SubscriberOptions struct {
	MaxConcurrentMessages int
}

type Subscriber struct {
	useCase application.DeliverProcessedMessagePort
	sem     chan struct{} // Semaphore: buffered channel with capacity = max concurrent
}

func NewSubscriber(useCase application.DeliverProcessedMessagePort, opts SubscriberOptions) *Subscriber {
	maxConcurrent := opts.MaxConcurrentMessages
	if maxConcurrent <= 0 {
		maxConcurrent = 50 // default
	}

	logrus.Infof("Sender semaphore configured: max %d concurrent messages", maxConcurrent)

	return &Subscriber{
		useCase: useCase,
		sem:     make(chan struct{}, maxConcurrent),
	}
}

func (s *Subscriber) Subscribe(js nats.JetStreamContext) error {
	_, err := js.Subscribe("messages.processed", func(msg *nats.Msg) {
		// Semaphore: acquire slot (blocks if at capacity)
		select {
		case s.sem <- struct{}{}:
			// Slot acquired, process asynchronously
			go s.sendMessageAsync(msg)
		case <-time.After(100 * time.Millisecond):
			// Timeout acquiring slot (shouldn't happen with proper sizing)
			logrus.Warnf("Timeout acquiring semaphore slot, rejecting message")
			_ = msg.Nak()
		}
	}, nats.Durable("sender"), nats.ManualAck())
	return err
}

// sendMessageAsync handles the actual message sending with semaphore slot acquired
func (s *Subscriber) sendMessageAsync(msg *nats.Msg) {
	// Release semaphore slot when done (guaranteed)
	defer func() { <-s.sem }()

	// Update metrics
	senderConcurrentMessages.Inc()
	defer senderConcurrentMessages.Dec()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var processed domain.MessageProcessed
	if err := json.Unmarshal(msg.Data, &processed); err != nil {
		logrus.Errorf("invalid processed message payload: %v", err)
		senderMessagesSent.WithLabelValues("error").Inc()
		_ = msg.Ack()
		return
	}

	if err := s.useCase.Execute(ctx, processed); err != nil {
		logrus.Errorf("delivery failed for external_id=%s: %v", processed.ExternalID, err)
		senderMessagesSent.WithLabelValues("error").Inc()
		_ = msg.Nak()
		return
	}

	logrus.Debugf("Sent message: external_id=%s, channel=%s", processed.ExternalID, processed.Channel)
	senderMessagesSent.WithLabelValues("success").Inc()
	_ = msg.Ack()
}
