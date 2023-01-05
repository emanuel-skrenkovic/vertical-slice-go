package product

import (
	"net/http"

	"github.com/eskrenkovic/mediator-go"
)

type ProductsEndpointHandler struct {
	m *mediator.Mediator
}

func NewProductsEndpointHandler(m *mediator.Mediator) *ProductsEndpointHandler {
	return &ProductsEndpointHandler{m}
}

func (h *ProductsEndpointHandler) HandleCreateProduct(w http.ResponseWriter, r *http.Request) {
}
