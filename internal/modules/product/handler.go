package product

import (
	"encoding/json"
	"net/http"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"
	"github.com/go-chi/chi"
	"github.com/google/uuid"

	"github.com/eskrenkovic/mediator-go"
)

type ProductsEndpointHandler struct {
	m *mediator.Mediator
}

func NewProductsEndpointHandler(m *mediator.Mediator) *ProductsEndpointHandler {
	return &ProductsEndpointHandler{m}
}

func (h *ProductsEndpointHandler) HandleCreateProduct(w http.ResponseWriter, r *http.Request) {
	var command CreateProductCommand
	if err := json.NewDecoder(r.Body).Decode(&command.Product); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if _, err := mediator.Send[CreateProductCommand, core.Unit](h.m, r.Context(), command); err != nil {
		if commandErr, ok := err.(core.CommandError); ok {
			w.WriteHeader(commandErr.StatusCode)
			bytes, err := json.Marshal(commandErr)
			if err != nil {
				// WTF do you do here?
			}
			w.Write(bytes)
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *ProductsEndpointHandler) HandleGetProduct(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "product_id")

	product, err := mediator.Send[GetProductQuery, Product](
		h.m,
		r.Context(),
		GetProductQuery{ProductID: uuid.MustParse(productID)})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bytes, err := json.Marshal(product)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Write(bytes)
}
