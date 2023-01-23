package config

import (
	"path"

	"github.com/eskrenkovic/vertical-slice-go/internal/env"
)

const (
	PortEnv        = "PORT"
	DatabaseUrlEnv = "DATABASE_URL"
	RootPathEnv    = "ROOT_PATH"
)

type Config struct {
	Port           int
	DatabaseURL    string
	MigrationsPath string
}

func Load() (Config, error) {
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
		Port:           port,
		DatabaseURL:    dbURL,
		MigrationsPath: migrationsPath,
	}, nil
}
