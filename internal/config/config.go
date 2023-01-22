package config

import (
	"path"

	"github.com/eskrenkovic/vertical-slice-go/internal/env"
)

const (
	DatabaseUrlEnv = "DATABASE_URL"
	RootPathEnv    = "ROOT_PATH"
)

type Config struct {
	DatabaseURL    string
	MigrationsPath string
}

func Load() (Config, error) {
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
		DatabaseURL:    dbURL,
		MigrationsPath: migrationsPath,
	}, nil
}
