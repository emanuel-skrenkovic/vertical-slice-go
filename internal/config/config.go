package config

import (
	"github.com/eskrenkovic/vertical-slice-go/internal/env"

	"github.com/joho/godotenv"
)

const (
	DatabaseUrlEnv = "DATABASE_URL"
)

type Config struct {
	DatabaseURL string
}

func Load(path string) (Config, error) {
	if err := godotenv.Load(path); err != nil {
		return Config{}, err
	}

	dbURL, err := env.GetString(DatabaseUrlEnv)
	if err != nil {
		return Config{}, err
	}

	return Config{
		DatabaseURL: dbURL,
	}, nil
}
