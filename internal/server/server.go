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

	"github.com/eskrenkovic/vertical-slice-go/internal/config"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/auth"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/commands"
	authcommands "github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/commands"
	authdomain "github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/domain"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"
	gamesessioncommands "github.com/eskrenkovic/vertical-slice-go/internal/modules/game-session/commands"
	gamesessiondomain "github.com/eskrenkovic/vertical-slice-go/internal/modules/game-session/domain"
	gamesessionqueries "github.com/eskrenkovic/vertical-slice-go/internal/modules/game-session/queries"

	"github.com/eskrenkovic/mediator-go"
	"github.com/eskrenkovic/migrate-go"
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

	server := http.Server{
		Addr: net.JoinHostPort("", "8080"),
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

	r := router{middleware: []httpMiddleware{
		baseContextMiddleware(baseCtx),
		core.CorrelationIDHTTPMiddleware,
	}}

	// http

	r.register("GET /game-sessions", gamesessionqueries.HandleGetOwnedSessions, auth.AuthenticationMiddleware(db))
	r.register("POST /game-sessions", gamesessioncommands.HandleCreateGameSession, auth.AuthenticationMiddleware(db))

	r.register("POST /game-sessions/{id}/invitations", gamesessioncommands.HandleCreateSessionInvitation, auth.AuthenticationMiddleware(db))

	r.register("PUT /game-sessions/{id}/actions/close", gamesessioncommands.HandleCloseSession, auth.AuthenticationMiddleware(db))
	r.register("PUT /game-sessions/{id}/actions/join", gamesessioncommands.HandleJoinSession, auth.AuthenticationMiddleware(db))

	r.register("POST /auth/login", authcommands.HandleLogin)
	r.register("POST /auth/logout", authcommands.HandleLogout)

	r.register("POST /auth/registrations", authcommands.HandleRegistration)
	r.register("POST /auth/registrations/actions/confirm", authcommands.HandleVerifyRegistration)
	r.register("POST /auth/registrations/actions/publish-confirmation-emails", authcommands.HandlePublishConfirmationEmails)
	r.register("POST /auth/registrations/actions/send-activation-code", authcommands.HandleReSendConfirmationEmail)

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

type httpMiddleware func(http.HandlerFunc) http.HandlerFunc

type router struct {
	middleware []httpMiddleware
}

func (r *router) register(pattern string, handler http.HandlerFunc, middleware ...httpMiddleware) {
	h := handler

	allMiddleware := append(r.middleware, middleware...)

	for i := len(allMiddleware) - 1; i >= 0; i-- {
		h = allMiddleware[i](h)
	}

	http.HandleFunc(pattern, h)
}

func baseContextMiddleware(baseCtx context.Context) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			baseCtx := baseCtx

			if v, ok := ctx.Value(http.ServerContextKey).(*http.Server); ok {
				baseCtx = context.WithValue(baseCtx, http.ServerContextKey, v)
			}

			if v, ok := ctx.Value(http.LocalAddrContextKey).(net.Addr); ok {
				baseCtx = context.WithValue(baseCtx, http.LocalAddrContextKey, v)
			}

			next.ServeHTTP(w, r.WithContext(baseCtx))
		}
	}
}
