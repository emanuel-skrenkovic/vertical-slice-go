package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/eskrenkovic/vertical-slice-go/internal/config"
	"github.com/eskrenkovic/vertical-slice-go/internal/server"
	"github.com/eskrenkovic/vertical-slice-go/internal/test"
	"go.uber.org/zap"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

type IntegrationTestFixture struct {
	client  *http.Client
	baseURL string
	db      *sqlx.DB
}

var fixture IntegrationTestFixture

func TestMain(m *testing.M) {
	args := os.Args

	if len(args) < 2 {
		log.Fatal("root path is required")
	}
	rootPath := args[len(args)-1]
	os.Setenv(config.RootPathEnv, rootPath)

	localConfigPath := path.Join(rootPath, "config.local.env")
	if _, err := os.Stat(localConfigPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			f, err := os.Create(localConfigPath)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			if _, err := f.Write([]byte("SKIP_INFRASTRUCTURE=false")); err != nil {
				log.Fatal(err)
			}
		}
	}

	if err := godotenv.Load(path.Join(rootPath, "config.local.env")); err != nil {
		log.Fatal(err)
	}

	if err := godotenv.Load(path.Join(rootPath, "config.env")); err != nil {
		log.Fatal(err)
	}

	config, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Replace logger with nop to avoid excess clutter during test runs.
	config.Logger = zap.NewNop()

	fixture, err := test.NewLocalTestFixture(path.Join(rootPath, "docker-compose.yml"), config.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	if err := fixture.Start(); err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := fixture.Stop(); err != nil {
			log.Fatal(err)
		}
	}()

	if err := initFixture(config); err != nil {
		log.Fatal(err)
	}

	server, err := server.NewHTTPServer(config)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := server.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	_ = m.Run()
}

func initFixture(config config.Config) error {
	fixture.client = &http.Client{}

	url := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", "localhost", config.Port),
	}
	fixture.baseURL = url.String()

	db, err := sqlx.Connect("postgres", config.DatabaseURL)
	if err != nil {
		return err
	}

	fixture.db = db

	return nil
}
