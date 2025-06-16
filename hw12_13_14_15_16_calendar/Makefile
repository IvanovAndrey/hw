BIN := "./bin/calendar"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar

run: build
	$(BIN) -config ./configs/config.toml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN) version

test:
	go test -race ./internal/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin  v1.60.2

lint: install-lint-deps
	golangci-lint run

DB_HOST     ?= 127.0.0.1
DB_PORT     ?= 7432
DB_USER     ?= postgres
DB_PASSWORD ?= postgres
DB_NAME     ?= postgres
DB_SSLMODE  ?= disable
DB_SCHEMA   ?= calendar,public
GOOSE_DIR   ?= ./migrations

DSN = "host=$(DB_HOST) port=$(DB_PORT) user=$(DB_USER) password=$(DB_PASSWORD) dbname=$(DB_NAME) search_path=$(DB_SCHEMA) sslmode=$(DB_SSLMODE)"

goose-install:
	go install github.com/pressly/goose/v3/cmd/goose@latest

migrate: goose-install
	goose -dir $(GOOSE_DIR) postgres $(DSN) up -v

.PHONY: build run build-img run-img version test lint migrate goose-install
