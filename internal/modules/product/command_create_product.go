package product

import (
	"context"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"
)

type CreateProductCommand struct {
	Product Product
}

type CreateProductHandler struct {
	repository *ProductRepository
}

func NewCreateProductHandler(repository *ProductRepository) *CreateProductHandler {
	return &CreateProductHandler{repository}
}

func (h *CreateProductHandler) Handle(ctx context.Context, request CreateProductCommand) (core.CommandResponse, error) {
	panic("TODO: implement CreateProductHandler.Handle")
}
