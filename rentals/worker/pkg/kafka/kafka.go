package kafka

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/Shopify/sarama"
	"github.com/okteto/movies/pkg/database"
)

// ConsumerGroupHandler implements sarama.ConsumerGroupHandler
type ConsumerGroupHandler struct {
	ctx          context.Context
	MessageCount int
	divertKey    string
	db           *sql.DB
	cg           sarama.ConsumerGroup
}

func NewConsumerGroup(ctx context.Context, namespace string, divertKey string, addrs []string, db *sql.DB) (*ConsumerGroupHandler, error) {
	// Create consumer group ID with namespace suffix
	consumerGroupID := fmt.Sprintf("movies-worker-group-%s", namespace)

	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	// Enable manual commit - we'll commit only after successful API calls
	config.Consumer.Offsets.AutoCommit.Enable = false

	consumerGroup, err := sarama.NewConsumerGroup(addrs, consumerGroupID, config)
	if err != nil {
		return nil, err
	}

	handler := &ConsumerGroupHandler{
		ctx:          ctx,
		MessageCount: 0,
		divertKey:    divertKey,
		db:           db,
		cg:           consumerGroup,
	}

	return handler, nil
}

func (c *ConsumerGroupHandler) Close() {
	c.cg.Close()
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) Consume(topics []string) error {
	return h.cg.Consume(h.ctx, topics, h)
}

/*func (c *ConsumerGroupHandler) shouldProcessMessage(baggage string) bool {
	// Extract okteto-divert value from baggage
	divertValue := extractOktetoDivertFromBaggage(baggage)

	// Rule 1: If message has okteto-divert key, process only if value matches environment variable
	if divertValue != "" {
		return divertValue == c.divertKey
	}

	// Rule 2: If message doesn't have okteto-divert key, process only if environment variable is empty
	return c.divertKey == ""

	// Rule 3: If this doesn't belong to anybody else, the 'shared' should get it
}*/

// extractOktetoDivertFromBaggage parses baggage string and extracts okteto-divert value
/*func extractOktetoDivertFromBaggage(baggage string) string {
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
}*/

// extractBaggageHeader extracts the baggage header value from Kafka message headers
/*func extractBaggageHeader(headers []*sarama.RecordHeader) string {
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
}*/

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		// Extract baggage header once for the message
		//baggageHeader := extractBaggageHeader(message.Headers)

		// Check if we should process this message based on divert logic
		//if !h.shouldProcessMessage(baggageHeader) {
		//	log.Printf("Not processing message, it belongs to a diverted worker")
		//	continue
		//}

		h.MessageCount++

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
