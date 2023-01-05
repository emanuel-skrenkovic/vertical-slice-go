package server

import (
	"log"
	"net"
	"net/http"

	"github.com/eskrenkovic/mediator-go"
	"github.com/eskrenkovic/vertical-slice-go/internal/config"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/product"
	productCommands "github.com/eskrenkovic/vertical-slice-go/internal/modules/product/commands"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

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
	server *http.Server
}

func NewHTTPServer(config config.Config) (Server, error) {
	router := chi.NewRouter()
	server := http.Server{
		Addr:    net.JoinHostPort("", "8080"),
		Handler: router,
	}

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

	productEndpointHandler := product.NewProductsEndpointHandler(m)

	router.Group(func(r chi.Router) {
		r.Use(middleware.StripSlashes)
		r.Use(middleware.RequestID)
		r.Use(middleware.Logger)

		router.Route("/products", func(r chi.Router) {
			r.Post("/", productEndpointHandler.HandleCreateProduct)
		})
	})

	return &HTTPServer{
		server: &server,
	}, nil
}

func (s *HTTPServer) Start() error {
	if err := s.server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

	return nil
}

func (s *HTTPServer) Stop() error {
	return s.server.Close()
}
