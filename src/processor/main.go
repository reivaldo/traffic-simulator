package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/yourusername/traffic-simulator/processor/internal/bootstrap"
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
	dbURL := viper.GetString("db_url")
	if dbURL == "" {
		dbURL = "postgres://traffic:traffic@postgres:5432/traffic?sslmode=disable"
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		logrus.Fatalf("NATS connection failed: %v", err)
	}
	defer nc.Close()
	js, _ := nc.JetStream()

	dbpool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		logrus.Fatalf("DB connection failed: %v", err)
	}
	defer dbpool.Close()

	// Read semaphore configuration
	maxConcurrent := viper.GetInt("processor.max_concurrent_messages")
	if maxConcurrent <= 0 {
		maxConcurrent = 50 // default
	}

	handler, err := bootstrap.Build(js, dbpool, maxConcurrent)
	if err != nil {
		logrus.Fatalf("NATS subscribe failed: %v", err)
	}

	port := viper.GetString("port")
	if port == "" {
		port = "8083"
	}
	logrus.Infof("Processor listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
