package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"database/sql"

	"github.com/Shopify/sarama"
	"github.com/okteto/movies/pkg/database"
)

var (
	topic             = kingpin.Flag("topic", "Topic name").Default("rentals").String()
	messageCountStart = kingpin.Flag("messageCountStart", "Message counter start from:").Int()
)

// ConsumerGroupHandler implements sarama.ConsumerGroupHandler
type ConsumerGroupHandler struct {
	messageCount *int
	divertKey    string
	db           *sql.DB
}

func main() {
	kingpin.Parse()

	// Get Kubernetes namespace from environment variable
	namespace := os.Getenv("KUBERNETES_NAMESPACE")
	if namespace == "" {
		namespace = "default"
		log.Printf("KUBERNETES_NAMESPACE not set, using default: %s", namespace)
	}

	// Get Kubernetes namespace from environment variable
	divertKey := os.Getenv("OKTETO_DIVERTED_ENVIRONMENT")
	if divertKey == "" {
		divertKey = ""
		log.Printf("OKTETO_DIVERTED_ENVIRONMENT not set, using default: %s", divertKey)
	}

	db, err := database.Open()
	if err != nil {
		log.Panic(err)
	}

	if err := database.LoadData(db); err != nil {
		log.Panic(err)
	}

	// Create consumer group ID with namespace suffix
	consumerGroupID := fmt.Sprintf("movies-worker-group-%s", namespace)
	log.Printf("Starting worker with consumer group ID: %s", consumerGroupID)

	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	// Enable manual commit - we'll commit only after successful API calls
	config.Consumer.Offsets.AutoCommit.Enable = false

	consumerGroup, err := sarama.NewConsumerGroup([]string{"kafka:9092"}, consumerGroupID, config)
	if err != nil {
		log.Panic(err)
	}
	defer consumerGroup.Close()

	handler := &ConsumerGroupHandler{
		messageCount: messageCountStart,
		divertKey:    divertKey,
		db:           db,
	}

	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			// Consume from both topics
			topics := []string{"rentals", "returns"}
			if err := consumerGroup.Consume(ctx, topics, handler); err != nil {
				log.Printf("Error from consumer: %v", err)
			}
			// Check if context was cancelled
			if ctx.Err() != nil {
				return
			}
		}
	}()

	<-signals
	fmt.Println("Interrupt is detected")
	cancel()
	wg.Wait()
	log.Println("Processed", *messageCountStart, "messages")
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *ConsumerGroupHandler) shouldProcessMessage(baggage string) bool {
	// Extract okteto-divert value from baggage
	divertValue := extractOktetoDivertFromBaggage(baggage)

	// Rule 1: If message has okteto-divert key, process only if value matches environment variable
	if divertValue != "" {
		return divertValue == c.divertKey
	}

	// Rule 2: If message doesn't have okteto-divert key, process only if environment variable is empty
	return c.divertKey == ""

	// Rule 3: If this doesn't belong to anybody else, the 'shared' should get it
}

// extractOktetoDivertFromBaggage parses baggage string and extracts okteto-divert value
func extractOktetoDivertFromBaggage(baggage string) string {
	if baggage == "" {
		return ""
	}

	// Parse baggage format: "key1=value1,key2=value2,..."
	pairs := strings.Split(baggage, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(kv) == 2 && strings.TrimSpace(kv[0]) == "okteto-divert" {
			return strings.TrimSpace(kv[1])
		}
	}

	return ""
}

// extractBaggageHeader extracts the baggage header value from Kafka message headers
func extractBaggageHeader(headers []*sarama.RecordHeader) string {
	for _, header := range headers {
		if string(header.Key) == "baggage" {
			baggageValue := string(header.Value)
			if baggageValue != "" {
				fmt.Printf("Baggage header found in Kafka message: %s\n", baggageValue)
			}
			return baggageValue
		}
	}
	return ""
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		// Extract baggage header once for the message
		baggageHeader := extractBaggageHeader(message.Headers)

		// Check if we should process this message based on divert logic
		if !h.shouldProcessMessage(baggageHeader) {
			log.Printf("Not processing message, it belongs to a diverted worker")
			continue
		}

		*h.messageCount++

		// Determine message type based on topic
		if message.Topic == "rentals" {
			if !h.processRentalMessage(string(message.Key), string(message.Value)) {
				// Don't commit if processing failed
				log.Printf("Failed to process rental message, will retry on next poll")
				continue
			}
		} else if message.Topic == "returns" {
			if !h.processReturnMessage(string(message.Value)) {
				// Don't commit if processing failed
				log.Printf("Failed to process return message, will retry on next poll")
				continue
			}
		}

		// Only mark message as consumed if processing was successful
		session.MarkMessage(message, "")
		// Commit the offset immediately after successful processing
		session.Commit()
	}
	return nil
}

// processRentalMessage handles rental messages and returns true if successful
func (h *ConsumerGroupHandler) processRentalMessage(movieID string, priceStr string) bool {
	fmt.Printf("Received message: movies %s price %s\n", movieID, priceStr)

	if err := database.CreateOrUpdateRental(h.db, movieID, priceStr); err != nil {
		log.Printf("Error processing the rental request: %v", err)
		return false
	}

	fmt.Printf("Successfully created/updated rental: %s - message committed\n", movieID)
	return true
}

// processReturnMessage handles return messages and returns true if successful
func (h *ConsumerGroupHandler) processReturnMessage(catalogID string) bool {
	fmt.Printf("Received return message: catalogID %s\n", catalogID)

	if err := database.DeleteRental(h.db, catalogID); err != nil {
		log.Printf("Error processing the delete rental request: %v", err)
		return false
	}

	fmt.Printf("Successfully deleted rental: %s - message committed\n", catalogID)
	return true
}
