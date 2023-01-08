.PHONY: build
build:
	go build cmd/api/main.go

run:
	go run cmd/api/main.go $(shell pwd)

test:
	go test -v ./pkg/sql-migrations/... -args $(shell pwd)
