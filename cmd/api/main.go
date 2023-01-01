package main

import (
	"log"
	"os"
	"path"

	"github.com/eskrenkovic/mediator-go"
	"github.com/eskrenkovic/vertical-slice-go/internal/config"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/product"
	productCommands "github.com/eskrenkovic/vertical-slice-go/internal/modules/product/commands"

	"github.com/jmoiron/sqlx"
)

func main() {
	rootPath := os.Args[1]
	if rootPath == "" {
		log.Fatal("root directoy path is empty")
	}

	config, err := config.Load(path.Join(rootPath, "config.env"))
	if err != nil {
		log.Fatal(err)
	}

	db, err := sqlx.Connect("postgres", config.DatabaseURL)
	if err != nil {
		log.Fatalln(err)
	}

	productRepository := product.NewProductRepository(db)

	m := mediator.NewMediator()

	createProductHandler := productCommands.NewCreateProductHandler(productRepository)
	err = mediator.RegisterRequestHandler[productCommands.CreateProductCommand, core.CommandResponse](
		m,
		createProductHandler,
	)
	if err != nil {
		log.Fatal(err)
	}
}
