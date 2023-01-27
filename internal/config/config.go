package config

import (
	"path"

	"github.com/eskrenkovic/vertical-slice-go/internal/env"
	"go.uber.org/zap"
)

const (
	PortEnv        = "PORT"
	DatabaseUrlEnv = "DATABASE_URL"
	RootPathEnv    = "ROOT_PATH"
)

type Config struct {
	Logger *zap.Logger

	Port           int
	DatabaseURL    string
	MigrationsPath string
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

	return Config{
		Logger:         logger,
		Port:           port,
		DatabaseURL:    dbURL,
		MigrationsPath: migrationsPath,
	}, nil
}
