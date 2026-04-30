package bootstrap

import (
	"net/http"

	"github.com/nats-io/nats.go"
	"github.com/yourusername/traffic-simulator/sender/internal/application"
	natsinfra "github.com/yourusername/traffic-simulator/sender/internal/infrastructure/nats"
	eventsiface "github.com/yourusername/traffic-simulator/sender/internal/interfaces/events"
	httpiface "github.com/yourusername/traffic-simulator/sender/internal/interfaces/http"
)

func Build(js nats.JetStreamContext, maxConcurrent int) (http.Handler, error) {
	publisher := natsinfra.NewDeliveryPublisher(js)
	useCase := application.NewDeliverProcessedMessage(publisher)
	subscriber := eventsiface.NewSubscriber(useCase, eventsiface.SubscriberOptions{
		MaxConcurrentMessages: maxConcurrent,
	})
	if err := subscriber.Subscribe(js); err != nil {
		return nil, err
	}
	return httpiface.NewRouter(), nil
}
