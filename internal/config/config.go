package config

import (
	"net/url"
	"path"

	"github.com/eskrenkovic/vertical-slice-go/internal/env"
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

	port, err := env.GetInt(PortEnv)
	if err != nil {
		return Config{}, err
	}

	dbURL, err := env.GetString(DatabaseUrlEnv)
	if err != nil {
		return Config{}, err
	}

	rootPath, err := env.GetString(RootPathEnv)
	if err != nil {
		return Config{}, err
	}

	migrationsPath := path.Join(rootPath, "db", "migrations")

	emailServerURL, err := env.GetURL(EmailServerHostEnv)
	if err != nil {
		return Config{}, err
	}

	emailServerUsername, err := env.GetString(EmailServerUsernameEnv)
	if err != nil {
		return Config{}, err
	}

	emailServerPassword, err := env.GetString(EmailServerPasswordEnv)
	if err != nil {
		return Config{}, err
	}

	emailServerSender, err := env.GetString(EmailServerSenderEnv)
	if err != nil {
		return Config{}, err
	}

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
