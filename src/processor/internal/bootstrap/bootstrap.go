package bootstrap

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/yourusername/traffic-simulator/processor/internal/application"
	natsinfra "github.com/yourusername/traffic-simulator/processor/internal/infrastructure/nats"
	postgresinfra "github.com/yourusername/traffic-simulator/processor/internal/infrastructure/postgres"
	eventsiface "github.com/yourusername/traffic-simulator/processor/internal/interfaces/events"
	httpiface "github.com/yourusername/traffic-simulator/processor/internal/interfaces/http"
)

func Build(js nats.JetStreamContext, db *pgxpool.Pool, maxConcurrent int) (http.Handler, error) {
	repo := postgresinfra.NewMessageRepository(db)
	publisher := natsinfra.NewProcessedPublisher(js)
	useCase := application.NewProcessAcceptedMessage(repo, publisher)
	
	subscriber := eventsiface.NewSubscriber(useCase, eventsiface.SubscriberOptions{
		MaxConcurrentMessages: maxConcurrent,
	})
	if err := subscriber.Subscribe(js); err != nil {
		return nil, err
	}
	return httpiface.NewRouter(), nil
}
