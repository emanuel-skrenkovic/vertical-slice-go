package server

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/eskrenkovic/mediator-go"
	"github.com/eskrenkovic/vertical-slice-go/internal/config"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/product"

	"github.com/eskrenkovic/vertical-slice-go/internal/sql-migrations"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
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
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	baseCtx := context.Background()

	router := chi.NewRouter()
	server := http.Server{
		Addr:    net.JoinHostPort("", "8080"),
		Handler: handlerWithBaseContext(baseCtx, router),
	}

	db, err := sqlx.Connect("postgres", config.DatabaseURL)
	if err != nil {
		return nil, err
	}

	if err := sqlmigration.Run("db/migrations", config.DatabaseURL); err != nil {
		return nil, err
	}

	productRepository := product.NewProductRepository(db)

	requestLoggingBehavior := core.RequestLoggingBehavior{Logger: logger}
	handlerErrorLoggingBehavior := core.HandlerErrorLoggingBehavior{Logger: logger}
	requestValidationBehavior := core.RequestValidationBehavior{}

	m := mediator.NewMediator()
	m.RegisterPipelineBehavior(&requestLoggingBehavior)
	m.RegisterPipelineBehavior(&handlerErrorLoggingBehavior)
	m.RegisterPipelineBehavior(&requestValidationBehavior)

	// handler registration
	createProductHandler := product.NewCreateProductHandler(productRepository)
	err = mediator.RegisterRequestHandler[product.CreateProductCommand, product.CreateProductResponse](
		m, createProductHandler,
	)
	if err != nil {
		return nil, err
	}

	getProductHandler := product.NewGetProductQueryHandler(productRepository)
	err = mediator.RegisterRequestHandler[product.GetProductQuery, product.Product](m, getProductHandler)
	if err != nil {
		return nil, err
	}

	// http
	productEndpointHandler := product.NewProductsEndpointHandler(m)

	router.Group(func(r chi.Router) {
		router.Route("/products", func(r chi.Router) {
			r.Use(middleware.StripSlashes)
			r.Use(middleware.Logger)
			r.Use(middleware.RequestID)
			r.Use(core.CorrelationIDHTTPMiddleware)

			r.Post("/", productEndpointHandler.HandleCreateProduct)
			r.Get("/{product_id}", productEndpointHandler.HandleGetProduct)
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

func handlerWithBaseContext(baseCtx context.Context, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		baseCtx := baseCtx

		if v, ok := ctx.Value(http.ServerContextKey).(*http.Server); ok {
			baseCtx = context.WithValue(baseCtx, http.ServerContextKey, v)
		}

		if v, ok := ctx.Value(http.LocalAddrContextKey).(net.Addr); ok {
			baseCtx = context.WithValue(baseCtx, http.LocalAddrContextKey, v)
		}

		handler.ServeHTTP(w, r.WithContext(baseCtx))
	})
}
