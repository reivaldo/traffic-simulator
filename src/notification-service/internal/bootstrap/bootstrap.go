package bootstrap

import (
	"net/http"

	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpinfra "github.com/yourusername/traffic-simulator/notification-service/internal/infrastructure/http"
	"github.com/yourusername/traffic-simulator/notification-service/internal/application"
	eventsiface "github.com/yourusername/traffic-simulator/notification-service/internal/interfaces/events"
)

func Build(js nats.JetStreamContext, smsURL, emailURL, whatsappURL string) (http.Handler, error) {
	// Infrastructure: HTTP gateways
	gateways := map[string]application.ProviderGateway{
		"sms":      httpinfra.NewSMSGateway(smsURL),
		"email":    httpinfra.NewEmailGateway(emailURL),
		"whatsapp": httpinfra.NewWhatsAppGateway(whatsappURL),
	}

	// Application: fan-out orchestrator
	fanOut := application.NewSendToAllProviders(gateways)

	// Interfaces: event subscriber
	subscriber := eventsiface.NewSubscriber(fanOut)
	if err := subscriber.Subscribe(js); err != nil {
		return nil, err
	}

	// HTTP router (basic healthz/metrics)
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ready\n"))
	})
	mux.Handle("/metrics", promhttp.Handler())

	return mux, nil
}
