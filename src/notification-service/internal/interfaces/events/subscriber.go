package eventsiface

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/traffic-simulator/notification-service/internal/application"
	"github.com/yourusername/traffic-simulator/notification-service/internal/domain"
)

var (
	notificationsSent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notification_service_deliveries_total",
			Help: "Total notification deliveries attempted",
		},
		[]string{"status"},
	)
)

func init() {
	prometheus.MustRegister(notificationsSent)
}

// Subscriber consumes messages.sent and sends to all providers (fan-out)
type Subscriber struct {
	useCase *application.SendToAllProviders
}

func NewSubscriber(useCase *application.SendToAllProviders) *Subscriber {
	return &Subscriber{useCase: useCase}
}

func (s *Subscriber) Subscribe(js nats.JetStreamContext) error {
	_, err := js.Subscribe("messages.sent", func(msg *nats.Msg) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var notification domain.Notification
		if err := json.Unmarshal(msg.Data, &notification); err != nil {
			logrus.Errorf("invalid message payload: %v", err)
			_ = msg.Ack()
			return
		}

		// Fan-out to all providers
		results, err := s.useCase.Execute(ctx, notification)
		if err != nil {
			logrus.Errorf("all providers failed: %v", err)
			notificationsSent.WithLabelValues("error").Inc()
			_ = msg.Nak()
			return
		}

		// Log results
		for _, result := range results {
			if result.Error != nil {
				logrus.Warnf("Provider %s failed: %v (%.2fms)", result.Provider, result.Error, result.Duration)
			} else {
				logrus.Infof("Provider %s succeeded: %s (%.2fms)", result.Provider, result.Status, result.Duration)
			}
		}

		notificationsSent.WithLabelValues("success").Inc()
		_ = msg.Ack()
	}, nats.Durable("notification-service"), nats.ManualAck())
	return err
}
