.PHONY: build
build:
	go build cmd/api/main.go

run:
	go run cmd/api/main.go $(shell pwd)

.PHONY: test
test:
	go test -v -count=1 ./test/api... -args $(shell pwd)
	go test -v -count=1 ./test/sql-migrations/... -args $(shell pwd)

.PHONY: docker-build
docker-build:
	docker build -t vertical-slice-go -t registry.fly.io/vertical-slice-go:latest --build-arg github_token=$(GITHUB_TOKEN) --build-arg goprivate=$(GOPRIVATE) .

.PHONY: infra-up
infra-up:
	[ -f config.local.env ] || echo "SKIP_INFRASTRUCTURE=false" > config.local.env
	sed -i 's/SKIP_INFRASTRUCTURE=false/SKIP_INFRASTRUCTURE=true/' config.local.env
	docker-compose up -d

.PHONY: infra-down
infra-down:
	[ -f config.local.env ] || echo "SKIP_INFRASTRUCTURE=false" > config.local.env
	sed -i 's/SKIP_INFRASTRUCTURE=true/SKIP_INFRASTRUCTURE=false/' config.local.env
	docker-compose down --remove-orphans
