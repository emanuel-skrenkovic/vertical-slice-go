package server

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"strings"

	"github.com/eskrenkovic/mediator-go"
	"github.com/eskrenkovic/migrate-go"
	"github.com/eskrenkovic/vertical-slice-go/internal/config"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/auth"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/commands"
	authcommands "github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/commands"
	authdomain "github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/domain"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"
	gamesessioncommands "github.com/eskrenkovic/vertical-slice-go/internal/modules/game-session/commands"
	gamesessiondomain "github.com/eskrenkovic/vertical-slice-go/internal/modules/game-session/domain"
	gamesessionqueries "github.com/eskrenkovic/vertical-slice-go/internal/modules/game-session/queries"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/lib/pq"
)

type Server interface {
	Start() error
	Stop() error
}

var _ Server = &HTTPServer{}

// HTTPServer acts as the composition root for an application.
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

	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		return nil, err
	}

	if err := migrate.Run(baseCtx, db, config.MigrationsPath); err != nil {
		return nil, err
	}

	requestLoggingBehavior := core.RequestLoggingBehavior{Logger: config.Logger}
	handlerErrorLoggingBehavior := core.HandlerErrorLoggingBehavior{Logger: config.Logger}
	requestValidationBehavior := core.RequestValidationBehavior{}

	mediator.RegisterPipelineBehavior(&requestLoggingBehavior)
	mediator.RegisterPipelineBehavior(&handlerErrorLoggingBehavior)
	mediator.RegisterPipelineBehavior(&requestValidationBehavior)

	// handler registration

	// game-session

	createGameSessionHandler := gamesessioncommands.NewCreateSessionCommandHandler(db)
	err = mediator.RegisterRequestHandler[gamesessioncommands.CreateSessionCommand, gamesessioncommands.CreateSessionResponse](
		createGameSessionHandler,
	)
	if err != nil {
		return nil, err
	}

	closeSessionHandler := gamesessioncommands.NewCloseSessionCommandHandler(db)
	err = mediator.RegisterRequestHandler[gamesessioncommands.CloseSessionCommand, core.Unit](
		closeSessionHandler,
	)
	if err != nil {
		return nil, err
	}

	getOwnedSessionsHandler := gamesessionqueries.NewGetOwnedSessionsQueryHandler(db)
	err = mediator.RegisterRequestHandler[gamesessionqueries.GetOwnedSessionsQuery, []gamesessiondomain.Session](
		getOwnedSessionsHandler,
	)
	if err != nil {
		return nil, err
	}

	createSessionInvitationHandler := gamesessioncommands.NewCreateSessionInvitationCommandHandler(db)
	err = mediator.RegisterRequestHandler[gamesessioncommands.CreateSessionInvitationCommand, core.Unit](
		createSessionInvitationHandler,
	)
	if err != nil {
		return nil, err
	}

	// auth
	authHost := config.Email.Host.Host
	parts := strings.Split(authHost, ":")
	if len(parts) > 1 {
		authHost = parts[0]
	}

	smtpServerAuth := smtp.PlainAuth("", config.Email.Username, config.Email.Password, authHost)
	emailClient := core.NewEmailClient(config.Email.Host, smtpServerAuth)
	passwordHasher := authdomain.NewPasswordHasher(sha256.New)

	loginHandler := authcommands.NewLoginCommandHandler(db, *passwordHasher)
	err = mediator.RegisterRequestHandler[authcommands.LoginCommand, authdomain.Session](
		loginHandler,
	)
	if err != nil {
		return nil, err
	}

	registerHandler := authcommands.NewRegisterCommandHandler(db, *passwordHasher)
	err = mediator.RegisterRequestHandler[authcommands.RegisterCommand, core.Unit](
		registerHandler,
	)
	if err != nil {
		return nil, err
	}

	verifyRegistrationCommandHandler := authcommands.NewVerifyRegistrationCommandHandler(db)
	err = mediator.RegisterRequestHandler[authcommands.VerifyRegistrationCommand, core.Unit](
		verifyRegistrationCommandHandler,
	)
	if err != nil {
		return nil, err
	}

	processActivationCodesCommandHandler := authcommands.NewProcessActivationCodesCommandHandler(
		db,
		emailClient,
		commands.EmailConfiguration{Sender: config.Email.Sender},
	)
	err = mediator.RegisterRequestHandler[authcommands.ProcessActivationCodesCommand, core.Unit](
		processActivationCodesCommandHandler,
	)
	if err != nil {
		return nil, err
	}

	reSendActivationEmailCommandHandler := authcommands.NewReSendActivationEmailCommandHandler(
		db,
		emailClient,
		config.Email.Sender,
	)
	err = mediator.RegisterRequestHandler[authcommands.ReSendActivationEmailCommand, core.Unit](
		reSendActivationEmailCommandHandler,
	)
	if err != nil {
		return nil, err
	}

	// http

	router.Group(func(r chi.Router) {
		router.Route("/game-sessions", func(r chi.Router) {
			r.Use(middleware.StripSlashes)
			r.Use(middleware.RequestID)
			r.Use(core.CorrelationIDHTTPMiddleware)

			r.Use(auth.AuthenticationMiddleware(db))

			r.Get("/", gamesessionqueries.HandleGetOwnedSessions)
			r.Post("/", gamesessioncommands.HandleCreateGameSession)

			r.Post("/{id}/invitations", gamesessioncommands.HandleCreateSessionInvitation)

			r.Put("/{id}/actions/close", gamesessioncommands.HandleCloseSession)
			r.Put("/{id}/actions/join", gamesessioncommands.HandleJoinSession)
		})

		router.Route("/auth", func(r chi.Router) {
			r.Use(middleware.StripSlashes)
			r.Use(middleware.RequestID)
			r.Use(core.CorrelationIDHTTPMiddleware)

			r.Post("/login", authcommands.HandleLogin)
			r.Post("/logout", authcommands.HandleLogout)
			r.Post("/registrations", authcommands.HandleRegistration)
			r.Post("/registrations/actions/confirm", authcommands.HandleVerifyRegistration)
			r.Post("/registrations/actions/publish-confirmation-emails", authcommands.HandlePublishConfirmationEmails)
			r.Post("/registrations/actions/send-activation-code", authcommands.HandleReSendConfirmationEmail)
		})
	})

	return &HTTPServer{server: &server}, nil
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
