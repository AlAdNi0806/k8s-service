// internal/repository/postgres.go
package repository

import (
	"context"

	"order-service/internal/model"

	"github.com/go-pg/pg/v10"
)

type OrderRepository struct {
	db *pg.DB
}

func NewOrderRepository(db *pg.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *model.Order) error {
	_, err := r.db.Model(order).Insert()
	return err
}
