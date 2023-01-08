package product

import (
	"context"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"

	"github.com/google/uuid"
)

type GetProductQuery struct {
	ProductID uuid.UUID
}

type GetProductQueryHandler struct {
	repository *ProductRepository
}

func NewGetProductQueryHandler(repository *ProductRepository) *GetProductQueryHandler {
	return &GetProductQueryHandler{repository}
}

func (h *GetProductQueryHandler) Handle(ctx context.Context, request GetProductQuery) (Product, error) {
	product, err := h.repository.LoadProduct(ctx, request.ProductID)
	if err != nil {
		return Product{}, core.NewCommandError(500, err, "failed to load product")
	}

	return product, nil
}
