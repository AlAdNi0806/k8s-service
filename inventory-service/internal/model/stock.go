// internal/model/stock.go
package model

type Stock struct {
	ProductID int64 `pg:"product_id,pk"`
	Quantity  int   `pg:"quantity,notnull"`
}
