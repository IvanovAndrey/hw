BIN_CALENDAR := ./bin/calendar
BIN_SCHEDULER := ./bin/calendar_scheduler
BIN_SENDER := ./bin/calendar_sender
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o $(BIN_CALENDAR)  -ldflags "$(LDFLAGS)" ./cmd/calendar
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o $(BIN_SCHEDULER) -ldflags "$(LDFLAGS)" ./cmd/calendar_scheduler
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o $(BIN_SENDER)    -ldflags "$(LDFLAGS)" ./cmd/calendar_sender

run: build
	$(BIN_CALENDAR) --config=./configs/calendar_config.yaml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN_CALENDAR) version

test:
	go test -race ./internal/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin  v2.1.6

lint: install-lint-deps
	golangci-lint run

DB_HOST     ?= postgres
DB_PORT     ?= 5432
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

install:
	go get \
		github.com/envoyproxy/protoc-gen-validate@v1.2.1 \
		github.com/favadi/protoc-go-inject-tag@v1.4.0 \
	&& go install \
		google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.1 \
	&& go install \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.19.1 \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.19.1 \
	&& go install \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4.0

PROTO_PATH = proto

protogen:
	@rm -rf ${PROTO_PATH}
	@mkdir ${PROTO_PATH}
	@protoc api/*.proto \
			-I external \
			-I api \
			-I ${GOPATH}/pkg/mod/github.com/grpc-ecosystem/grpc-gateway/v2@v2.10.3 \
			-I${GOPATH}/pkg/mod/github.com/envoyproxy/protoc-gen-validate@v0.6.7 \
			--go_out=${PROTO_PATH} \
			--go_opt=paths=source_relative \
			--go-grpc_out=${PROTO_PATH} \
			--go-grpc_opt=paths=source_relative \
			--validate_out=lang=go:${PROTO_PATH}/ \
			--validate_opt=paths=source_relative \
			--grpc-gateway_out=allow_delete_body=true:${PROTO_PATH} \
			--grpc-gateway_opt paths=source_relative \
			--openapiv2_out=logtostderr=true,allow_delete_body=true,json_names_for_fields=false,allow_merge=true,merge_file_name=prodg_api_server:${PROTO_PATH}

.PHONY: build run build-img run-img version test lint migrate goose-install

DOCKER_COMPOSE := docker-compose

up:
	$(DOCKER_COMPOSE) up --build calendar calendar_scheduler calendar_sender postgres rabbitmq

down:
	$(DOCKER_COMPOSE) down -v

integration-tests:
	$(DOCKER_COMPOSE) down -v
	$(DOCKER_COMPOSE) up --abort-on-container-exit integration-tests
	$(DOCKER_COMPOSE) down -v

prepare:
	$(DOCKER_COMPOSE) up --build postgres rabbitmq migrator

run_services:
	$(DOCKER_COMPOSE) up --build calendar calendar_scheduler calendar_sender