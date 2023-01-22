package main

import (
	"log"
	"os"
	"path"
	"testing"

	"github.com/eskrenkovic/vertical-slice-go/internal/test"
)

func TestMain(m *testing.M) {
	args := os.Args

	if len(args) < 2 {
		log.Fatal("root path is required")
	}
	rootPath := args[len(args)-1]

	fixture, err := test.NewLocalTestFixture(path.Join(rootPath, "docker-compose.yml"))
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

	m.Run()
}
