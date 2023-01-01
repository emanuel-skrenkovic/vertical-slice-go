package product

import (
	"context"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/product"
)

type CreateProductCommand struct {
	Product product.Product
}

type CreateProductHandler struct {
	repository *product.ProductRepository
}

func NewCreateProductHandler(repository *product.ProductRepository) *CreateProductHandler {
	return &CreateProductHandler{repository}
}

func (h *CreateProductHandler) Handle(ctx context.Context, request CreateProductCommand) (core.CommandResponse, error) {
	panic("TODO: implement CreateProductHandler.Handle")
}
