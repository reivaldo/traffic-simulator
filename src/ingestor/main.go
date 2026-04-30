package main

import (
	"log"
	"net/http"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/yourusername/traffic-simulator/ingestor/internal/bootstrap"
)

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	_ = viper.ReadInConfig()

	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	natsURL := viper.GetString("nats_url")
	if natsURL == "" {
		natsURL = "nats://nats:4222"
	}
	nc, err := nats.Connect(natsURL)
	if err != nil {
		logrus.Fatalf("NATS connection failed: %v", err)
	}
	defer nc.Close()

	handler := bootstrap.BuildHTTPHandler(nc)

	port := viper.GetString("port")
	if port == "" {
		port = "8082"
	}
	logrus.Infof("Ingestor listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
