package product

import (
	"net/http"
	"path"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"

	"github.com/eskrenkovic/mediator-go"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type ProductsEndpointHandler struct {
	m *mediator.Mediator
}

func NewProductsEndpointHandler(m *mediator.Mediator) *ProductsEndpointHandler {
	return &ProductsEndpointHandler{m}
}

func (h *ProductsEndpointHandler) HandleCreateProduct(w http.ResponseWriter, r *http.Request) {
	command, err := core.RequestBody[CreateProductCommand](r)
	if err != nil {
		core.WriteBadRequest(w, r, err)
		return
	}

	response, err := mediator.Send[CreateProductCommand, CreateProductResponse](h.m, r.Context(), command)
	if err != nil {
		// TODO: don't like this at all. Needs to be a simple function call or a decorator solution.
		statusCode := 500
		if commandErr, ok := err.(core.CommandError); ok {
			statusCode = commandErr.StatusCode
		}
		core.WriteResponse(w, r, statusCode, err)
		return
	}

	location := path.Join(r.Host, "products", response.ProductID.String())
	core.WriteCreated(w, r, location)
}

func (h *ProductsEndpointHandler) HandleGetProduct(w http.ResponseWriter, r *http.Request) {
	productIDParam := chi.URLParam(r, "product_id")
	productID, err := uuid.Parse(productIDParam)
	if err != nil {
		core.WriteBadRequest(w, r, err)
		return
	}

	query := GetProductQuery{ProductID: productID}
	product, err := mediator.Send[GetProductQuery, Product](h.m, r.Context(), query)
	if err != nil {
		statusCode := 500
		if commandErr, ok := err.(core.CommandError); ok {
			statusCode = commandErr.StatusCode
		}
		core.WriteResponse(w, r, statusCode, err)
		return
	}

	core.WriteOK(w, r, product)
}
