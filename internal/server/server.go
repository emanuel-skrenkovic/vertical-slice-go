package server

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/eskrenkovic/mediator-go"
	"github.com/eskrenkovic/vertical-slice-go/internal/config"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"


	gamesession "github.com/eskrenkovic/vertical-slice-go/internal/modules/game-session"
	gamesessiondomain "github.com/eskrenkovic/vertical-slice-go/internal/modules/game-session/domain"
	gamesessioncommands "github.com/eskrenkovic/vertical-slice-go/internal/modules/game-session/commands"
	gamesessionqueries "github.com/eskrenkovic/vertical-slice-go/internal/modules/game-session/queries"

	auth "github.com/eskrenkovic/vertical-slice-go/internal/modules/auth"
	authdomain "github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/domain"
	authcommands "github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/commands"


	sqlmigration "github.com/eskrenkovic/vertical-slice-go/internal/sql-migrations"

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

	if err := sqlmigration.Run(config.MigrationsPath, config.DatabaseURL); err != nil {
		return nil, err
	}

	requestLoggingBehavior := core.RequestLoggingBehavior{Logger: config.Logger}
	handlerErrorLoggingBehavior := core.HandlerErrorLoggingBehavior{Logger: config.Logger}
	requestValidationBehavior := core.RequestValidationBehavior{}

	m := mediator.NewMediator()
	m.RegisterPipelineBehavior(&requestLoggingBehavior)
	m.RegisterPipelineBehavior(&handlerErrorLoggingBehavior)
	m.RegisterPipelineBehavior(&requestValidationBehavior)

	// handler registration

	// game-session

	createGameSessionHandler := gamesessioncommands.NewCreateSessionCommandHandler(db)
	err = mediator.RegisterRequestHandler[gamesessioncommands.CreateSessionCommand, gamesessioncommands.CreateSessionResponse](
		m,
		createGameSessionHandler,
	)
	if err != nil {
		return nil, err
	}

	closeSessionHandler := gamesessioncommands.NewCloseSessionCommandHandler(db)
	err = mediator.RegisterRequestHandler[gamesessioncommands.CloseSessionCommand, core.Unit](
		m,
		closeSessionHandler,
	)
	if err != nil {
		return nil, err
	}

	getOwnedSessionsHandler := gamesessionqueries.NewGetOwnedSessionsQueryHandler(db)
	err = mediator.RegisterRequestHandler[gamesessionqueries.GetOwnedSessionsQuery, []gamesessiondomain.Session](
		m,
		getOwnedSessionsHandler,
	)
	if err != nil {
		return nil, err
	}

	// auth
	passwordHasher := authdomain.NewSHA256PasswordHasher()

	loginHandler := authcommands.NewLoginCommandHandler(db, passwordHasher)
	err = mediator.RegisterRequestHandler[authcommands.LoginCommand, core.Unit](
		m,
		loginHandler,
	)
	if err != nil {
		return nil, err
	}


	registerHandler := authcommands.NewRegisterCommandHandler(db, passwordHasher)
	err = mediator.RegisterRequestHandler[authcommands.RegisterCommand, core.Unit](
		m,
		registerHandler,
	)
	if err != nil {
		return nil, err
	}

	verifyRegistrationCommandHandler := authcommands.NewVerifyRegistrationCommandHandler(db)
	err = mediator.RegisterRequestHandler[authcommands.VerifyRegistrationCommand, core.Unit](
		m,
		verifyRegistrationCommandHandler,
	)
	if err != nil {
		return nil, err
	}

	// http

	// Game sessions
	gameSessionEndpointHandler := gamesession.NewGameSessionHTTPHandler(m)

	// auth
	authEndpointHandler := auth.NewAuthHTTPHandler(m)

	router.Group(func(r chi.Router) {
		router.Route("/game-sessions", func(r chi.Router) {
			r.Use(middleware.StripSlashes)
			r.Use(middleware.Logger)
			r.Use(middleware.RequestID)
			r.Use(core.CorrelationIDHTTPMiddleware)

			r.Get("/", gameSessionEndpointHandler.HandleGetOwnedSessions)
			r.Post("/", gameSessionEndpointHandler.HandleCreateGameSession)
			r.Put("/{id}/actions/close", gameSessionEndpointHandler.HandleCloseSession)
		})

		router.Route("/auth", func(r chi.Router) {
			r.Use(middleware.StripSlashes)
			r.Use(middleware.Logger)
			r.Use(middleware.RequestID)
			r.Use(core.CorrelationIDHTTPMiddleware)

			r.Post("/login", authEndpointHandler.HandleLogin)
			r.Post("/logout", authEndpointHandler.HandleLogout)
			r.Post("/registration", authEndpointHandler.HandleRegistration)
			r.Post("/registration/actions/confirm", authEndpointHandler.HandleVerifyRegistration)
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
