package env

import (
	"errors"
	"fmt"
	"os"
)

var (
	ErrNotFound         = errors.New("environment variable with key not found")
	ErrConversionFailed = errors.New("failed to convert environment variable with key to value")
)

func errNotFound(key string) error {
	return fmt.Errorf("key: %s: %w", key, ErrNotFound)
}

func errConversionFailed(key string, typeName string) error {
	return fmt.Errorf("key: %s type: %s: %w", key, typeName, ErrConversionFailed)
}

func GetStringOrDefault(key string, defaultVal string) string {
	if val, found := os.LookupEnv(key); found {
		return val
	}

	return defaultVal
}

func GetString(key string) (string, error) {
	if val, found := os.LookupEnv(key); found {
		return val, nil
	}

	return "", errNotFound(key)
}
