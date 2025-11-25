package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/okteto/movies/pkg/database"
	"github.com/okteto/movies/pkg/kafka"
)

func getEnv(key, value string) string {
	v := os.Getenv(key)
	if v == "" {
		return value
	}

	return v
}

func main() {
	kingpin.Parse()
	ctx, cancel := context.WithCancel(context.Background())

	// Get Kubernetes namespace from environment variable
	namespace := getEnv("KUBERNETES_NAMESPACE", "default")

	// Get Kubernetes namespace from environment variable
	divertKey := getEnv("OKTETO_DIVERTED_ENVIRONMENT", "")

	// Kafka address
	addrs := getEnv("KAFKA_ADDRESS", "kafka:9092")

	db, err := database.Open()
	if err != nil {
		log.Panic(err)
	}

	database.Ping(db)

	if err := database.LoadData(db); err != nil {
		log.Panic(err)
	}

	handler, err := kafka.NewConsumerGroup(ctx, namespace, divertKey, []string{addrs}, db)
	if err != nil {
		log.Panic(err)
	}

	defer handler.Close()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			// Consume from both topics
			topics := []string{"rentals", "returns"}
			if err := handler.Consume(topics); err != nil {
				log.Printf("Error from consumer: %v", err)
			}
			// Check if context was cancelled
			if ctx.Err() != nil {
				return
			}
		}
	}()

	<-signals
	log.Println("Interrupt is detected")
	cancel()
	wg.Wait()
	log.Println("Processed", handler.MessageCount, "messages")
}
