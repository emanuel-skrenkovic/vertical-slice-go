package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/eskrenkovic/vertical-slice-go/internal/config"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/tests"
	"github.com/eskrenkovic/vertical-slice-go/internal/server"

	"github.com/joho/godotenv"
)

type IntegrationTestFixture struct {
	client  *http.Client
	baseURL string
	db      *sql.DB
}

var fixture = IntegrationTestFixture{}

func TestMain(m *testing.M) {
	rootPath := "../../"
	if err := os.Setenv(config.RootPathEnv, rootPath); err != nil {
		log.Fatal(err)
	}

	localConfigPath := path.Join(rootPath, "config.local.env")
	if _, err := os.Stat(localConfigPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			f, err := os.Create(localConfigPath)
			if err != nil {
				log.Fatal(err)
			}
			defer func() {
				if err := f.Close(); err != nil {
					log.Fatal(err)
				}
			}()

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

	conf, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	conf.Logger = zap.NewNop()

	pgPort := nat.Port(fmt.Sprintf("%d", 5432))
	mailhogPort := nat.Port(fmt.Sprintf("%d", 8025))

	waitStrategies := map[string]wait.Strategy{
		"vsg-postgres": wait.ForSQL(pgPort, "postgres", func(string, nat.Port) string { return conf.DatabaseURL }),
		"vsg-mailhog":  wait.ForHTTP("").WithPort(mailhogPort),
	}

	ctx := context.Background()

	composePath := path.Join(rootPath, "docker-compose.yml")
	fmt.Println(composePath)
	f, err := tests.NewLocalTestFixture(composePath, waitStrategies)

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("unrecovarable error occurred: %+v", err)
		}
	}()

	defer func() {
		if err := f.Stop(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	if err := f.Start(ctx); err != nil {
		log.Fatal(err)
	}

	if err := initFixture(conf); err != nil {
		log.Fatal(err)
	}

	srv, err := server.NewHTTPServer(conf)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	_ = m.Run()
}

func initFixture(config config.Config) error {
	fixture.client = &http.Client{}

	u := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", "localhost", config.Port),
	}
	fixture.baseURL = u.String()

	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		return err
	}

	fixture.db = db

	return nil
}
