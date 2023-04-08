package env

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strconv"
)

var (
	ErrNotFound         = errors.New("environment variable with key not found")
	ErrConversionFailed = errors.New("failed to convert environment variable with key to value")
)

func errNotFound(key string) error {
	return fmt.Errorf("key: %s: %w", key, ErrNotFound)
}

func errConversionFailed(key string, typeName string, err error) error {
	return fmt.Errorf("key: %s type: %s: %w", key, typeName, ErrConversionFailed)
}

func MustGetString(key string) string {
	if val, found := os.LookupEnv(key); found {
		return val
	}

	panic(errNotFound(key))
}

func MustGetInt(key string) int {
	envVal, found := os.LookupEnv(key)
	if !found {
		panic(errNotFound(key))
	}

	val, err := strconv.Atoi(envVal)
	if err != nil {
		panic(errConversionFailed(key, reflect.TypeOf(val).Name(), err))
	}

	return val
}

func MustGetURL(key string) *url.URL {
	val, found := os.LookupEnv(key)
	if !found {
		panic(errNotFound(key))
	}

	u, err := url.Parse(val)
	if err != nil {
		panic(errConversionFailed(key, reflect.TypeOf(u).Name(), err))
	}

	return u
}
