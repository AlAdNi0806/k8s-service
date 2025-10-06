// internal/model/order.go
package model

import "time"

type Order struct {
	ID        int64     `pg:"id,pk"`
	UserID    int64     `pg:"user_id,notnull"`
	ProductID int64     `pg:"product_id,notnull"`
	Quantity  int       `pg:"quantity,notnull"`
	Status    string    `pg:"status,default:'pending'"`
	CreatedAt time.Time `pg:"created_at,default:now()"`
}
