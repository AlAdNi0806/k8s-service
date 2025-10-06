// internal/repository/postgres.go
package repository

import (
	"context"
	"fmt"

	"inventory-service/internal/model"

	"github.com/go-pg/pg/v10"
)

type InventoryRepository struct {
	db *pg.DB
}

func NewInventoryRepository(db *pg.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

func (r *InventoryRepository) GetStock(ctx context.Context, productID int64) (*model.Stock, error) {
	stock := &model.Stock{ProductID: productID}
	err := r.db.Model(stock).Where("product_id = ?", productID).Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil // товар ещё не заведён
		}
		return nil, err
	}
	return stock, nil
}

func (r *InventoryRepository) DeductStock(ctx context.Context, productID int64, quantity int) error {
	// Используем UPDATE с проверкой, чтобы избежать отрицательных остатков
	res, err := r.db.ExecContext(ctx, `
        UPDATE stock
        SET quantity = quantity - ?
        WHERE product_id = ? AND quantity >= ?`,
		quantity, productID, quantity)

	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("insufficient stock for product %d", productID)
	}

	return nil
}

func (r *InventoryRepository) EnsureStock(ctx context.Context, productID int64, initialQty int) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO stock (product_id, quantity)
        VALUES (?, ?)
        ON CONFLICT (product_id) DO NOTHING`,
		productID, initialQty)
	return err
}
