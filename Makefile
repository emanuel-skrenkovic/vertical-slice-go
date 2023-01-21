.PHONY: build
build:
	go build cmd/api/main.go
	go build pkg/env/env.go

run:
	go run cmd/api/main.go $(shell pwd)

.PHONY: test
test:
	go test -v -count=1 ./pkg/sql-migrations/... -args $(shell pwd)
