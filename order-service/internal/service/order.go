// internal/service/order.go
package service

import (
	"context"
	"encoding/json"
	"log"

	"order-service/internal/model"
	"order-service/internal/repository"
	"order-service/internal/utils"

	"github.com/segmentio/kafka-go"
)

type OrderService struct {
	orderRepo   *repository.OrderRepository
	kafkaWriter *kafka.Writer
}

type OrderEvent struct {
	OrderID   int64 `json:"order_id"`
	UserID    int64 `json:"user_id"`
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

func NewOrderService(orderRepo *repository.OrderRepository, kafkaWriter *kafka.Writer) *OrderService {
	return &OrderService{orderRepo: orderRepo, kafkaWriter: kafkaWriter}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID, productID int64, quantity int) error {
	logger := utils.NewHelperLogger("order-service.service.create-order")

	order := &model.Order{
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
		Status:    "pending",
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		logger.LogError(ctx, "Failed to create order in database", err,
			log.KeyValue{Key: "user_id", Value: log.Int64Value(userID)},
			log.KeyValue{Key: "product_id", Value: log.Int64Value(productID)},
		)
		return err
	}

	// Публикуем событие в Kafka
	event := OrderEvent{
		OrderID:   order.ID,
		UserID:    order.UserID,
		ProductID: order.ProductID,
		Quantity:  order.Quantity,
	}

	payload, _ := json.Marshal(event)
	err := s.kafkaWriter.WriteMessages(ctx, kafka.Message{
		Topic: "order.created",
		Value: payload,
	})
	if err != nil {
		logger.LogError(ctx, "Failed to publish to Kafka", err,
			log.KeyValue{Key: "order_id", Value: log.Int64Value(order.ID)},
		)
		// Можно добавить retry или dead-letter queue в продакшене
	} else {
		logger.LogInfo(ctx, "Order created and published to Kafka",
			log.KeyValue{Key: "order_id", Value: log.Int64Value(order.ID)},
		)
	}

	return nil
}

func (s *OrderService) ValidateToken(token string) (int64, error) {
	return utils.ValidateToken(token)
}
