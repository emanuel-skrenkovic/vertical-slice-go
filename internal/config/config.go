package config

import (
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/env"
	"net/url"
	"path"

	"go.uber.org/zap"
)

const (
	PortEnv        = "PORT"
	DatabaseUrlEnv = "DATABASE_URL"
	RootPathEnv    = "ROOT_PATH"

	EmailServerHostEnv     = "EMAIL_SERVER_HOST"
	EmailServerUsernameEnv = "EMAIL_SERVER_USERNAME"
	EmailServerPasswordEnv = "EMAIL_SERVER_PASSWORD"
	EmailServerSenderEnv   = "EMAIL_SERVER_SENDER"
)

type EmailConfiguration struct {
	Host     *url.URL
	Username string
	Password string
	Sender   string
}

type Config struct {
	Logger *zap.Logger

	Port           int
	DatabaseURL    string
	MigrationsPath string

	Email EmailConfiguration
}

func Load() (Config, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return Config{}, err
	}

	port := env.MustGetInt(PortEnv)
	dbURL := env.MustGetString(DatabaseUrlEnv)

	rootPath := env.MustGetString(RootPathEnv)

	emailServerURL := env.MustGetURL(EmailServerHostEnv)
	emailServerUsername := env.MustGetString(EmailServerUsernameEnv)
	emailServerPassword := env.MustGetString(EmailServerPasswordEnv)
	emailServerSender := env.MustGetString(EmailServerSenderEnv)

	migrationsPath := path.Join(rootPath, "db", "migrations")

	return Config{
		Logger:         logger,
		Port:           port,
		DatabaseURL:    dbURL,
		MigrationsPath: migrationsPath,
		Email: EmailConfiguration{
			Host:     emailServerURL,
			Username: emailServerUsername,
			Password: emailServerPassword,
			Sender:   emailServerSender,
		},
	}, nil
}
