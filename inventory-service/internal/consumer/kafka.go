// internal/consumer/kafka.go
package consumer

import (
	"context"
	"log"

	"inventory-service/internal/service"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader *kafka.Reader
	svc    *service.InventoryService
}

func NewKafkaConsumer(brokers []string, groupID, topic string, svc *service.InventoryService) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	return &KafkaConsumer{reader: reader, svc: svc}
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	log.Println("Starting Kafka consumer...")
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		if err := c.svc.HandleOrderEvent(ctx, msg.Value); err != nil {
			log.Printf("Error handling message: %v", err)
			// В продакшене: логика повтора или DLQ
		} else {
			log.Printf("Message processed: offset=%d", msg.Offset)
		}
	}
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
