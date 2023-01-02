package main

import (
	"log"
	"os"
	"path"

	"github.com/eskrenkovic/vertical-slice-go/internal/config"
	"github.com/eskrenkovic/vertical-slice-go/internal/server"
)

// TODO: will need to move this to a separate struct to
// be able to Start()/Stop() in integration tests.
func main() {
	rootPath := os.Args[1]
	if rootPath == "" {
		log.Fatal("root directoy path is empty")
	}

	config, err := config.Load(path.Join(rootPath, "config.env"))
	if err != nil {
		log.Fatal(err)
	}

	server, err := server.NewHTTPServer(config)
	if err != nil {
		log.Fatal(err)
	}

	if err = server.Start(); err != nil {
		log.Fatal(err)
	}

	// TODO: this doesn't work.
	defer func() {
		if err := server.Stop(); err != nil {
			log.Fatal(err)
		}
	}()
}
