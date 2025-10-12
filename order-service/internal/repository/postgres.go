// internal/repository/postgres.go
package repository

import (
	"context"
	"database/sql"
	"time"

	"order-service/internal/model"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *model.Order) error {
	// Insert the order and return the generated ID (MySQL syntax)
	query := `
		INSERT INTO orders (user_id, product_id, quantity, status, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		order.UserID,
		order.ProductID,
		order.Quantity,
		order.Status,
		time.Now(),
	)
	if err != nil {
		return err
	}

	// Get the last inserted ID for MySQL
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	order.ID = id

	return nil
}
