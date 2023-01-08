package product

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ProductRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db}
}

func (r *ProductRepository) LoadProduct(ctx context.Context, id uuid.UUID) (Product, error) {
	const query = `
		SELECT *
		FROM product
		WHERE id = $1;`

	var product Product
	return product, r.db.GetContext(ctx, &product, query, id)
}

// TODO: should this be upsert? I kind of want to separate
// update and insert.
func (r *ProductRepository) SaveProduct(ctx context.Context, product Product) error {
	const script = `
		INSERT INTO product (id, sku, name, description, price)
		VALUES(:id, :sku, :name, :description, :price)
		ON CONFLICT (id)
		DO
		UPDATE
		SET sku=:sku, name=:name, description=:description, price=:price
		WHERE product.id=:id;`

	_, err := r.db.NamedExecContext(ctx, script, product)
	return err
}
