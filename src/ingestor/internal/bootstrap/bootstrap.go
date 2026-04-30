package bootstrap

import (
	"net/http"

	"github.com/nats-io/nats.go"
	"github.com/yourusername/traffic-simulator/ingestor/internal/application"
	natsinfra "github.com/yourusername/traffic-simulator/ingestor/internal/infrastructure/nats"
	httpiface "github.com/yourusername/traffic-simulator/ingestor/internal/interfaces/http"
)

func BuildHTTPHandler(nc *nats.Conn) http.Handler {
	publisher := natsinfra.NewIntentPublisher(nc)
	acceptUC := application.NewAcceptMessageIntent(publisher)
	handler := httpiface.NewHandler(acceptUC)
	return handler.Router()
}
