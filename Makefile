SERVER_SRC_DIR=./cmd/server
SERVER_BINARY_NAME=server
BUILD_DIR=./bin
MIGRATIONS_DIR=./migrations

CGO_ENABLED=0
GOARCH=amd64
GOOS=linux

.PHONY: all
all: tidy fmt lint run/server

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: mock
mock:
	mockery

.PHONY: test/unit
test/unit:
	go test -v -cover -race ./...

.PHONY: build/server
build/server:
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOARCH=$(GOARCH) GOOS=$(GOOS) go build -o $(BUILD_DIR)/$(SERVER_BINARY_NAME) $(SERVER_SRC_DIR)

.PHONY: run/server
run/server:
	go run $(SERVER_SRC_DIR)/*

.PHONY: clean
clean:
	rm -fr $(BUILD_DIR)
	go clean -testcache

.PHONY: migrations/create
migrations/create:
	ifndef MIGRATION_NAME
		$(error MIGRATION_NAME is not defined)
	endif
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(MIGRATION_NAME)

.PHONY: migrations/up
migrations/up:
	@if [ -z "$(DATABASE_DSN)" ]; then \
		echo "Error: DATABASE_DSN is not defined"; \
		exit 1; \
	fi; \
	migrate -database $(DATABASE_DSN) -path $(MIGRATIONS_DIR) up

.PHONY: migrations/down
migrations/down:
	@if [ -z "$(DATABASE_DSN)" ]; then \
		echo "Error: DATABASE_DSN is not defined"; \
		exit 1; \
	fi; \
	migrate -database $(DATABASE_DSN) -path $(MIGRATIONS_DIR) down -all

.PHONY: ci
ci: tidy fmt build/server lint test/unit clean

.PHONY: git/push
git/push: ci
	@if [ -z "$(BRANCH)" ]; then \
		BRANCH=$(shell git rev-parse --abbrev-ref HEAD); \
	fi; \
	git push -u origin $(BRANCH)
