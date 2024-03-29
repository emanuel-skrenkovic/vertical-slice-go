package main

import (
	"log"
	"os"
	"path"

	"github.com/eskrenkovic/vertical-slice-go/internal/config"
	"github.com/eskrenkovic/vertical-slice-go/internal/server"

	"github.com/joho/godotenv"
)

// TODO: will need to move this to a separate struct to
// be able to Start()/Stop() in integration tests.
func main() {
	if len(os.Args) > 1 {
		rootPath := os.Args[1]
		if rootPath == "" {
			log.Fatal("root directory path is empty")
		}

		if err := godotenv.Load(path.Join(rootPath, "config.env")); err != nil {
			log.Fatal(err)
		}
	}

	conf, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	srv, err := server.NewHTTPServer(conf)
	if err != nil {
		log.Fatal(err)
	}

	if err = srv.Start(); err != nil {
		log.Fatal(err)
	}

	// TODO: this doesn't work.
	defer func() {
		if err := srv.Stop(); err != nil {
			log.Fatal(err)
		}
	}()
}
