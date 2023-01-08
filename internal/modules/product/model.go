package product

import (
	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `db:"id"`
	SKU         string    `db:"sku"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Price       int       `db:"price"`
}

type CreateProductModel struct {
	SKU         string    `json:"sku"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       int       `json:"price"`
}
