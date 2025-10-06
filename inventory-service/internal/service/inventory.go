// internal/service/inventory.go
package service

import (
	"context"
	"encoding/json"
	"log"

	"inventory-service/internal/repository"
)

type OrderEvent struct {
	OrderID   int64 `json:"order_id"`
	UserID    int64 `json:"user_id"`
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

type InventoryService struct {
	repo *repository.InventoryRepository
}

func NewInventoryService(repo *repository.InventoryRepository) *InventoryService {
	return &InventoryService{repo: repo}
}

func (s *InventoryService) HandleOrderEvent(ctx context.Context, msg []byte) error {
	var event OrderEvent
	if err := json.Unmarshal(msg, &event); err != nil {
		return err
	}

	log.Printf("Processing order %d: product=%d, qty=%d", event.OrderID, event.ProductID, event.Quantity)

	// Убедимся, что товар существует (для демо можно предварительно заполнить)
	// В реальной системе — проверка наличия в каталоге

	if err := s.repo.DeductStock(ctx, event.ProductID, event.Quantity); err != nil {
		log.Printf("Failed to deduct stock: %v", err)
		// В продакшене: отправить в DLQ или повторить
		return err
	}

	log.Printf("Stock deducted for product %d", event.ProductID)
	return nil
}
