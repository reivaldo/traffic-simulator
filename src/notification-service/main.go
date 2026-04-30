package main

import (
	"log"
	"net/http"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/yourusername/traffic-simulator/notification-service/internal/bootstrap"
)

func main() {
	// Config loader
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	_ = viper.ReadInConfig()

	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	// NATS connection
	natsURL := viper.GetString("nats_url")
	if natsURL == "" {
		natsURL = "nats://nats:4222"
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		logrus.Fatalf("failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		logrus.Fatalf("failed to get JetStream context: %v", err)
	}

	logrus.Infof("Connected to NATS: %s", natsURL)

	// Provider URLs
	smsURL := viper.GetString("sms_url")
	if smsURL == "" {
		smsURL = "http://localhost:9000"
	}

	emailURL := viper.GetString("email_url")
	if emailURL == "" {
		emailURL = "http://localhost:9001"
	}

	whatsappURL := viper.GetString("whatsapp_url")
	if whatsappURL == "" {
		whatsappURL = "http://localhost:9002"
	}

	logrus.Infof("Provider URLs: SMS=%s, Email=%s, WhatsApp=%s", smsURL, emailURL, whatsappURL)

	// Bootstrap application
	handler, err := bootstrap.Build(js, smsURL, emailURL, whatsappURL)
	if err != nil {
		logrus.Fatalf("failed to bootstrap: %v", err)
	}

	port := viper.GetString("port")
	if port == "" {
		port = "8085"
	}

	logrus.Infof("Notification service listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
