package server

import (
	"log"
	"net/http"

	"github.com/eskrenkovic/mediator-go"
	"github.com/eskrenkovic/vertical-slice-go/internal/config"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/product"
	productCommands "github.com/eskrenkovic/vertical-slice-go/internal/modules/product/commands"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Server interface {
	Start() error
	Stop() error
}

var _ Server = &HTTPServer{}

// Acts as the composition root for an application.
type HTTPServer struct {
}

func NewHTTPServer(config config.Config) (Server, error) {
	db, err := sqlx.Connect("postgres", config.DatabaseURL)
	if err != nil {
		return nil, err
	}

	productRepository := product.NewProductRepository(db)

	m := mediator.NewMediator()

	createProductHandler := productCommands.NewCreateProductHandler(productRepository)
	err = mediator.RegisterRequestHandler[productCommands.CreateProductCommand, core.CommandResponse](
		m,
		createProductHandler,
	)
	if err != nil {
		return nil, err
	}

	return &HTTPServer{}, nil
}

func (s *HTTPServer) Start() error {
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

	return nil
}

func (s *HTTPServer) Stop() error {
	return nil
}
