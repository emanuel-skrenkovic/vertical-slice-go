.PHONY: build
build:
	go build cmd/api/main.go

run:
	go run cmd/api/main.go $(shell pwd)

.PHONY: test
test:
	go test -v -count=1 ./internal/sql-migrations/... -args $(shell pwd)
	go test -v -count=1 ./test/... -args $(shell pwd)

.PHONY: docker-build
docker-build:
	docker build -t vertical-slice-go -t registry.fly.io/vertical-slice-go:latest --build-arg github_token=$(GITHUB_TOKEN) --build-arg goprivate=$(GOPRIVATE) .
