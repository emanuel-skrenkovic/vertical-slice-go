package product

import (
	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID
	SKU         string
	Name        string
	Description string
	Price       int
}
