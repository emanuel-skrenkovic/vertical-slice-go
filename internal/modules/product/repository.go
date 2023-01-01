package product

import (
	"github.com/jmoiron/sqlx"
)

type ProductRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db}
}

func LoadProduct() (Product, error) {
	panic("TODO: implement ProductRepository.LoadProduct")
}

func SaveProduct(product Product) (error) {
	panic("TODO: implement ProductRepository.SaveProduct")
}
