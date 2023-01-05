.PHONY: build
build:
	go build cmd/api/main.go

run:
	go run cmd/api/main.go $(shell pwd)
